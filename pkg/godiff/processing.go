/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 *
 */
package godiff

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type GoDiffDataStruct struct {
	Origin          string
	Destination     string
	missingInOrg    []string
	missingPath     []string
	missingInDest   []string
	wrongTypeInOrg  []string
	wrongTypeInDest []string
	unmatchFile     []string
}

func init() {
	file, err := os.OpenFile("results.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Formatter = &logrus.TextFormatter{
			FullTimestamp:          false,
			DisableTimestamp:       true,
			ForceColors:            true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
		}
		multi := io.MultiWriter(os.Stdout, file)
		log.SetOutput(multi)
		// log.Out = file
		// log.SetOutput(os.Stdout)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func filesEqual(file1, file2 string) (bool, error) {
	/*
		Compare hashes of file1 and file2 and return a boolean:
		true if files are equal,
		false if not
	*/
	// Open file
	f1, err := os.Open(file1)
	if err != nil {
		log.Error("Failed to open: ", file1, " error: ", err)
		return false, err
	}
	defer f1.Close()
	f2, err := os.Open(file2)
	if err != nil {
		log.Error("Failed to open: ", file2, " error: ", err)
		return false, err
	}
	defer f2.Close()

	// Get the hash of the first file
	h1 := md5.New()
	if _, err := io.Copy(h1, f1); err != nil {
		log.Error("Failed to get hash for: ", file1, " error: ", err)
		return false, err
	}
	hash1 := hex.EncodeToString(h1.Sum(nil))

	// Get the hash of the second file
	h2 := md5.New()
	if _, err := io.Copy(h2, f2); err != nil {
		log.Error("Failed to get hash for: ", file1, " error: ", err)
		return false, err
	}
	hash2 := hex.EncodeToString(h2.Sum(nil))

	// Compare the hashes of the files
	if hash1 != hash2 {
		log.Warn("Files: ", file1, " and: ", file2, " are differents.")
		return false, nil
	}
	return true, nil
}

func checkFile(path1 string, path2 string) (bool, error) {
	/*
		Read file1 and file2 and return boolean if file content are strickly equal:
		true if files are equal, error nil
		false if there is a difference, error nil
		false and error not nil if an error occur
	*/
	statFile1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}
	statFile2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}
	if statFile1.IsDir() {
		return false, fmt.Errorf("Path: %s is a directoy", path1)
	}
	if statFile2.IsDir() {
		return false, fmt.Errorf("Path: %s is a directoy", path2)
	}

	file1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer file1.Close()
	file2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer file2.Close()

	buf1 := make([]byte, 1024)
	buf2 := make([]byte, 1024)

	for {
		n1, err1 := file1.Read(buf1)
		n2, err2 := file2.Read(buf2)
		if err1 != nil || err2 != nil || n1 != n2 {
			return false, nil
			break
		}
		if n1 == 0 {
			break
		}
		if string(buf1[:n1]) != string(buf2[:n2]) {
			return false, nil
			break
		}
	}
	return true, nil
}

func (p *GoDiffDataStruct) Process(dir1 string, dir2 string) error {
	/*
		Walk through the first directory and compare each files with the second directory:
		check by file hashes,
		if different, compare line by line.
		report and log results.
	*/
	// Walk through DIR 1
	filepath.Walk(dir1, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Get the corresponding file in the second directory
		relPath, _ := filepath.Rel(dir1, path)
		path2 := filepath.Join(dir2, relPath)
		file1, err := os.Stat(path)
		if err != nil {
			fmt.Errorf("Error in: %s, %s", path, err)
			return nil
		}
		file2, err := os.Stat(path2)
		if err != nil {
			if !stringInSlice(path, p.missingPath) {
				if file1.IsDir() {
					log.Info("Directory is missing: ", path, "\n")
					p.missingPath = append(p.missingPath, path)
					// Skip this dir if the current path is missing, no need to walk through all subdir
					return filepath.SkipDir
				} else {
					log.Warn("File is missing: ", path, "\n")
					p.missingPath = append(p.missingPath, path)
				}
			}
		} else {
			if file1.IsDir() && !file2.IsDir() {
				if !stringInSlice(path, p.wrongTypeInOrg) {
					log.Warn("File: ", path, "and: ", path2, " have different type (directory vs file)")
					p.wrongTypeInOrg = append(p.wrongTypeInOrg, path)
				}
			}
			if !file1.IsDir() && file2.IsDir() {
				if !stringInSlice(path, p.wrongTypeInDest) {
					log.Warn("File: ", path, "and: ", path2, " have different type (directory vs file)")
					p.wrongTypeInDest = append(p.wrongTypeInDest, path2)
				}
			}
			if !file1.IsDir() && !file2.IsDir() {
				check, err := filesEqual(path, path2)
				if err != nil {
					return err
				}
				if !check {
					// Compare the two files
					if !stringInSlice(path, p.unmatchFile) {
						p.unmatchFile = append(p.unmatchFile, path)
						report, err := CompareFiles(path, path2, false, true)
						if err != nil {
							return err
						}

						if report != nil {
							if !stringInSlice(path, p.unmatchFile) {
								p.unmatchFile = append(p.unmatchFile, path)
							}
							if !stringInSlice(path2, p.unmatchFile) {
								p.unmatchFile = append(p.unmatchFile, path2)
							}
						}
					}

				}

			}
		}
		return nil
	})
	return nil
}

func (p *GoDiffDataStruct) ProcessDirectories(reverse bool) error {
	// Compare origin vs destination
	log.Info("Start processing: ", p.Origin, " as source and: ", p.Destination, " as destination.")
	p.Process(p.Origin, p.Destination)
	if reverse {
		p.Process(p.Destination, p.Origin)
	}
	fmt.Printf("\n**** Report ****\n")
	if len(p.missingPath) > 0 {
		fmt.Printf("\n**** Missing files or directories ****\n")
		fmt.Println(strings.Join(p.missingPath, "\n"))
	}
	if len(p.unmatchFile) > 0 {
		fmt.Printf("\n**** Files with differences ****\n")
		fmt.Println(strings.Join(p.unmatchFile, "\n"))
	}
	if len(p.wrongTypeInOrg) > 0 {
		fmt.Printf("\n**** Different file type in origin ****\n")
		fmt.Println(strings.Join(p.wrongTypeInOrg, "\n"))
	}
	return nil
}
