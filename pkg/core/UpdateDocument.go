package core

// UpdateDocument inserts or updates a document in the zinc index
func (ind *Index) UpdateDocument(docID string, doc *map[string]interface{}, mintedID bool) error {
	bdoc, err := ind.BuildBlugeDocFromJSON(docID, doc)
	if err != nil {
		return err
	}

	// Finally update the document on disk
	writer := ind.Writer
	if !mintedID {
		err = writer.Update(bdoc.ID(), bdoc)
	} else {
		err = writer.Insert(bdoc)
	}
	// err = writer.Update(bdoc.ID(), bdoc)
	if err != nil {
		return err
	}

	return nil
}
