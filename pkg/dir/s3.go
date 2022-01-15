package dir

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/index"
	segment "github.com/blugelabs/bluge_segment_api"
	"github.com/rs/zerolog/log"
)

// GetS3Config returns a bluge config that will store index data in S3
// bucket: the S3 bucket to use
// indexName: the name of the index to use. It will be an s3 prefix (folder)
func GetS3Config(bucket, indexName string) bluge.Config {
	return bluge.DefaultConfigWithDirectory(func() index.Directory {
		return newS3Dir(bucket, indexName)
	})
}

type s3Dir struct {
	Bucket string
	Prefix string
	Client *s3.Client
}

// newS3Dir creates a new s3Dir instance which can be used to create s3 backed indexes
func newS3Dir(bucket, prefix string) index.Directory {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Print("Error loading AWS config: ", err)
	}
	client := s3.NewFromConfig(cfg)

	directory := &s3Dir{
		Bucket: bucket,
		Prefix: prefix,
		Client: client,
	}

	return directory
}

func (s *s3Dir) fileName(kind string, id uint64) string {
	return fmt.Sprintf("%012x", id) + kind
}

func (s *s3Dir) Setup(readOnly bool) error {
	return nil
}

// List the ids of all the items of the specified kind
// Items are returned in descending order by id
func (s *s3Dir) List(kind string) ([]uint64, error) {
	log.Print("List: s3 ListObjectsV2 call made for List: s3://", s.Bucket+"/"+s.Prefix)
	var itemList []uint64

	params := s3.ListObjectsV2Input{
		Bucket: &s.Bucket,
		Prefix: &s.Prefix,
	}

	val, err := s.Client.ListObjectsV2(context.Background(), &params)
	if err != nil {
		log.Print("List: failed to list objects: ", err.Error())
		return nil, err
	}

	for _, obj := range val.Contents {
		if filepath.Ext(*obj.Key) != kind {
			continue
		}

		stringID := filepath.Base(*obj.Key)
		stringID = stringID[:len(stringID)-len(kind)]

		parsedID, err := strconv.ParseUint(stringID, 16, 64)
		if err != nil {
			log.Print("List: failed to parse object id: ", err.Error())
			continue
		}

		itemList = append(itemList, parsedID)

	}

	return itemList, nil
}

// Load the specified items.
// Item data is accessible via the returned *segment.Data structure
// A io.Closer is returned, which must be called to release
// resources held by this open item.
// NOTE: care must be taken to handle a possible nil io.Closer
func (s *s3Dir) Load(kind string, id uint64) (*segment.Data, io.Closer, error) {
	key := s.Prefix + "/" + s.fileName(kind, id)
	goi := &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	}

	log.Print("Load: s3 GetObject call made. s3://", s.Bucket, "/", key)
	output, err := s.Client.GetObject(context.Background(), goi)
	if err != nil {
		log.Print("Load: failed to get object: s3://"+s.Bucket+"/"+key, err.Error())
		return nil, nil, err
	}

	data, err := ioutil.ReadAll(output.Body)
	if err != nil {
		log.Print("Load: failed to read object", err.Error())
		return nil, nil, err
	}

	return segment.NewDataBytes(data), nil, nil
}

// Persist a new item with data from the provided WriterTo
// Implementations should monitor the closeCh and return with error
// in the event it is closed before completion.
func (s *s3Dir) Persist(kind string, id uint64, w index.WriterTo, closeCh chan struct{}) error {
	var buf bytes.Buffer
	_, err := w.WriteTo(&buf, closeCh)
	if err != nil {
		log.Print("Persist: failed to write object to buffer: ", err.Error())
		return err
	}

	s3ObjectName := s.fileName(kind, id)
	path := filepath.Join(s.Prefix, s3ObjectName)
	params := s3.PutObjectInput{
		Bucket: &s.Bucket,
		Key:    &path,
		Body:   bytes.NewReader(buf.Bytes()),
	}

	ouput, err := s.Client.PutObject(context.Background(), &params)
	if err != nil {
		log.Print("Persist: failed to write object: ", err.Error())
		return err
	}

	log.Print("Persist: s3 object "+s.Bucket+"/"+path+" written. Its md5 hash is: ", *ouput.ETag) // TODO: compare md5 hashes here to ensure successful write

	return nil
}

// Remove the specified item
func (s *s3Dir) Remove(kind string, id uint64) error {
	objectToDelete := filepath.Join(s.Prefix, s.fileName(kind, id))
	ctx := context.Background()
	doi := &s3.DeleteObjectInput{
		Bucket: &s.Bucket,
		Key:    &objectToDelete,
	}

	log.Print("Remove: s3 DeleteObject call made s3://", s.Bucket, "/", objectToDelete)

	_, err := s.Client.DeleteObject(ctx, doi)
	if err != nil {
		log.Print("Remove: failed to delete object: s3://", s.Bucket, "/", objectToDelete, err.Error())
	}
	return nil
}

// Stats returns total number of items and their cumulative size
func (s *s3Dir) Stats() (objectCount, sizeOfObjects uint64) {
	log.Print("Stats: s3 ListObjectsV2 call made for Stats")

	ctx := context.Background()
	params := s3.ListObjectsV2Input{
		Bucket: &s.Bucket,
		Prefix: &s.Prefix,
	}

	log.Print("Stats: s3 ListObjectsV2 call made for Stats s3://", s.Bucket+"/"+s.Prefix)

	val, err := s.Client.ListObjectsV2(ctx, &params)
	if err != nil {
		log.Print("Stats: failed to list objects: ", err.Error())
		return 0, 0
	}

	for _, obj := range val.Contents {
		size := uint64(obj.Size)
		objectCount++
		sizeOfObjects += size
	}

	return objectCount, sizeOfObjects
}

// Sync ensures directory metadata itself has been committed
func (s *s3Dir) Sync() error {
	return nil
}

// Lock ensures this process has exclusive access to write in this directory
func (s *s3Dir) Lock() error {
	return nil
}

// Unlock releases the lock held on this directory
func (s *s3Dir) Unlock() error {
	return nil
}
