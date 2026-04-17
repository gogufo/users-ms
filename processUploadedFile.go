package main

import (
	. "github.com/gogufo/gufo-api-gateway/gufodao"
	"os"
)

// processUploadedFile contains business logic for handling fully received files.
// This function is called asynchronously after "end" meta is received.
func processUploadedFile(path string) {
	// EXAMPLES:
	// TODO: implement your real business logic here
	// 1. Parse data (PDF, CSV, JSON, ZIP etc)
	_, err := os.ReadFile(path)
	if err != nil {
		SetErrorLog("cannot read: " + err.Error())
		return
	}

	// 2. Convert to another format
	//extracted := extractMetadata(content)

	// 3. Forward to another microservice
	//callInventoryMS(extracted)

	// 4. Upload to S3
	//err = uploadToS3(path)

	// 5. Save metadata to DB
	//saveToDB(extracted)

	// 6. Trigger background processing pipeline
	//scheduleJob(path)
}
