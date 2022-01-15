package core

// UpdateDoc inserts or updates a document in the zinc index
func (ind *Index) UpdateDoc(docID string, doc *map[string]interface{}, mintedID bool) error {
	bdoc, err := ind.BuildBlugeDocFromJSON(docID, doc)
	if err != nil {
		return err
	}

	// Finally, update the document on disk
	if mintedID {
		return ind.Writer.Insert(bdoc)
	}

	return ind.Writer.Update(bdoc.ID(), bdoc)
}
