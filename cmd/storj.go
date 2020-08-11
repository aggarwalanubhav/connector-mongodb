// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"storj.io/uplink"
)

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
	SerializedAccess     string `json:"serializedAccess"`
	AllowDownload        string `json:"allowDownload"`
	AllowUpload          string `json:"allowUpload"`
	AllowList            string `json:"allowList"`
	AllowDelete          string `json:"allowDelete"`
	NotBefore            string `json:"notBefore"`
	NotAfter             string `json:"notAfter"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) ConfigStorj {

	var configStorj ConfigStorj
	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal("Could not load storj config file: ", err)
	}

	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configStorj); err != nil {
		log.Fatal(err)
	}

	// Close the file handle after reading from it.
	if err = fileHandle.Close(); err != nil {
		log.Fatal(err)
	}

	// Display storj configuration read from file.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)

	// Convert the upload path to standard form.
	checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Access Key\t: ", configStorj.SerializedAccess)
	return configStorj
}

// ShareAccess generates and prints the shareable serialized access
// as per the restrictions provided by the user.
func ShareAccess(access *uplink.Access, configStorj ConfigStorj) {

	allowDownload, _ := strconv.ParseBool(configStorj.AllowDownload)
	allowUpload, _ := strconv.ParseBool(configStorj.AllowUpload)
	allowList, _ := strconv.ParseBool(configStorj.AllowList)
	allowDelete, _ := strconv.ParseBool(configStorj.AllowDelete)
	notBefore, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotBefore)
	notAfter, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotAfter)

	permission := uplink.Permission{
		AllowDownload: allowDownload,
		AllowUpload:   allowUpload,
		AllowList:     allowList,
		AllowDelete:   allowDelete,
		NotBefore:     notBefore,
		NotAfter:      notAfter,
	}

	// Create shared access.
	sharedAccess, err := access.Share(permission)
	if err != nil {
		log.Fatal("Could not generate shared access: ", err)
	}

	// Generate restricted serialized access.
	serializedAccess, err := sharedAccess.Serialize()
	if err != nil {
		log.Fatal("Could not serialize shared access: ", err)
	}
	fmt.Println("Shareable sererialized access: ", serializedAccess)
}

// ConnectToStorj reads Storj configuration from given file
// and connects to the desired Storj network.
// It then reads data property from an external file.
func ConnectToStorj(fullFileName string, configStorj ConfigStorj, accesskey bool) (*uplink.Access, *uplink.Project) {

	var access *uplink.Access
	var cfg uplink.Config

	// Configure the UserAgent
	cfg.UserAgent = "MongoDB"
	ctx := context.Background()
	var err error

	if accesskey {
		fmt.Println("\nConnecting to Storj network using Serialized access.")
		// Generate access handle using serialized access.
		access, err = uplink.ParseAccess(configStorj.SerializedAccess)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("\nConnecting to Storj network.")
		// Generate access handle using API key, satellite url and encryption passphrase.
		access, err = cfg.RequestAccessWithPassphrase(ctx, configStorj.Satellite, configStorj.APIKey, configStorj.EncryptionPassphrase)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Open a new porject.
	project, err := cfg.OpenProject(ctx, access)
	if err != nil {
		log.Fatal(err)
	}
	defer project.Close()

	// Ensure the desired Bucket within the Project
	_, err = project.EnsureBucket(ctx, configStorj.Bucket)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to Storj network.")
	return access, project
}

// UploadData uploads the backup file to storj network.
func UploadData(project *uplink.Project, configStorj ConfigStorj, uploadFileName string, dbReader io.Reader, firstCollection string) {

	ctx := context.Background()

	// Create an upload handle for the first collection.
	upload, err := project.UploadObject(ctx, configStorj.Bucket, configStorj.UploadPath+uploadFileName+"/"+firstCollection+".bson", nil)
	if err != nil {
		log.Fatal("Could not initiate upload : ", err)
	}
	fmt.Printf("\nUploading %s to %s...\n", configStorj.UploadPath+uploadFileName+"/"+firstCollection+".bson", configStorj.Bucket)

	buf := make([]byte, 10485760)
	var err1 = io.ErrShortBuffer
	// Loop to upload and commit each collection one by one.
	for err1 != nil {
		_, err1 = io.CopyBuffer(upload, dbReader, buf)
		if err1 != nil && err1 != io.ErrShortBuffer {
			// Commit the current copied collection.
			err = upload.Commit()
			if err != nil {
				log.Fatal("Could not commit object upload : ", err)
			}
			// Create upload handle for the next collection to be uploaded.
			upload, err = project.UploadObject(ctx, configStorj.Bucket, configStorj.UploadPath+uploadFileName+"/"+currentCollection+".bson", nil)
			if err != nil {
				log.Fatal("Could not initiate upload : ", err)
			}
			fmt.Printf("Uploading %s to %s...\n", configStorj.UploadPath+uploadFileName+"/"+currentCollection+".bson", configStorj.Bucket)
		}
	}

	// Commit the upload after copying the last collection.
	err = upload.Commit()
	if err != nil {
		log.Fatal("Could not commit object upload : ", err)
	}
}

func findLatestBackup(project *uplink.Project, configStorj ConfigStorj, databaseName string) *uplink.Object {
	ctx := context.Background()
	// Object iterator to traverse all the back-ups of the specified database.
	objects := project.ListObjects(ctx, configStorj.Bucket, &uplink.ListObjectsOptions{Prefix: configStorj.UploadPath + databaseName + "/"})
	var latestBackup *uplink.Object
	i := 1
	// Loop to find the latest back-up of all the back-ups.
	for objects.Next() {
		item := objects.Item()
		if i == 1 && item.System.Created.Before(time.Now()) {
			latestBackup = item
		} else {
			if item.System.Created.After(latestBackup.System.Created) {
				latestBackup = item
			}
		}
		i++
	}

	return latestBackup
}

// RestoreData restores the latest backup correspoinding to the path provided
func RestoreData(project *uplink.Project, configStorj ConfigStorj, databaseName string, latest bool) {

	ctx := context.Background()
	var collections *uplink.ObjectIterator
	if latest {
		fmt.Printf("Restoring the latest backup of %s...\n", databaseName)
		checkSlash := databaseName[len(databaseName)-1:]
		if checkSlash == "/" {
			databaseName = databaseName[:len(databaseName)-1]
		}
		pathTokens := strings.Split(databaseName, "/")
		if len(pathTokens) > 1 {
			log.Fatal("Error: Invalid regular expression! It should only contain the patter of database name.\n")
		}
		latestBackup := findLatestBackup(project, configStorj, databaseName)
		collections = project.ListObjects(ctx, configStorj.Bucket, &uplink.ListObjectsOptions{Prefix: latestBackup.Key})
	} else {
		fmt.Printf("Restoring the backup of %s...\n", databaseName)
		// Convert the backup path to standard form
		checkSlash := databaseName[len(databaseName)-1:]
		if checkSlash != "/" {
			databaseName = databaseName + "/"
		}
		collections = project.ListObjects(ctx, configStorj.Bucket, &uplink.ListObjectsOptions{Prefix: configStorj.UploadPath + databaseName})
	}

	// Download all the collection back-up files corresponding to the back-up inside the ./dump folder.
	for collections.Next() {
		item := collections.Item()
		download, err := project.DownloadObject(ctx, configStorj.Bucket, item.Key, nil)
		if err != nil {
			log.Fatal(err)
		}
		// Read everything from the download stream
		receivedContents, err := ioutil.ReadAll(download)
		if err != nil {
			log.Fatal(err)
		}

		downloadFileName := filepath.Join("dump", filepath.Base(filepath.Dir(item.Key)), filepath.Base(item.Key))
		_ = os.MkdirAll(filepath.Dir(downloadFileName), 0644)
		err = ioutil.WriteFile(downloadFileName, receivedContents, 0600)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Latest backup of %s restored.\n", databaseName)
}

// MatchAndRestore finds the databases corresponding the pattern entered by the user
// and restores the latest backup of each matching database.
func MatchAndRestore(project *uplink.Project, configStorj ConfigStorj, matchPattern string, latest bool) {

	if !latest {
		log.Fatal("Error: match used without latest flag!")
	}
	//	var validDatabaseName = regexp.MustCompile(matchPattern)
	ctx := context.Background()
	databases := project.ListObjects(ctx, configStorj.Bucket, &uplink.ListObjectsOptions{Prefix: configStorj.UploadPath})
	for databases.Next() {
		item := databases.Item()
		matched, err := regexp.MatchString(matchPattern, item.Key)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			fmt.Println("Matching database: ", item.Key)
			RestoreData(project, configStorj, filepath.Base(item.Key), latest)
		}
	}
}