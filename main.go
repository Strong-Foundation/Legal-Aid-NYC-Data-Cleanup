package main // Main package

import (
	"bufio"
	"fmt"           // For printing output
	"io"            // For copying data
	"log"           // For logging errors and information
	"net/http"      // For making HTTP requests
	"net/url"       // For parsing input URLs
	"os"            // For file and directory operations
	"path"          // For manipulating file paths
	"path/filepath" // For working with file paths
	"regexp"        // For extracting ID and slug using regex
	"strings"       // For string manipulation
)

// extractFinalDocumentCloudURL converts documentcloud.org or embed.documentcloud.org links to final S3 links
func extractFinalDocumentCloudURL(input string) string {
	parsedURL, err := url.Parse(input) // Parse the input into a URL struct
	if err != nil {                    // If parsing fails
		return "" // Return empty string
	}

	// If it's already a direct S3 URL, return it
	if strings.Contains(parsedURL.Host, "s3.documentcloud.org") {
		return input // Already good, return as-is
	}

	// Regex to match both www and embed DocumentCloud URLs
	// Matches format: <domain>/documents/<docID>-<slug>
	re := regexp.MustCompile(`documentcloud\.org/documents/(\d+)-([\w\-]+)`) // Capture ID and slug

	// Run regex on the input URL
	matches := re.FindStringSubmatch(input)
	if len(matches) != 3 { // If match didn't find both parts
		return "" // Return empty string
	}

	docID := matches[1] // First captured group is the document ID
	slug := matches[2]  // Second captured group is the slug

	// Format the final S3 URL
	finalURL := fmt.Sprintf("https://s3.documentcloud.org/documents/%s/%s.pdf", docID, slug)

	return finalURL // Return the constructed final URL
}

// downloadPDF downloads a PDF from a direct S3 URL and saves it to the specified output directory.
// It skips downloading if the file already exists locally.
func downloadPDF(finalURL, outputDir string) {
	// Parse the URL to work with its path and file name
	parsedURL, err := url.Parse(finalURL)
	if err != nil {
		log.Printf("Invalid URL %q: %v", finalURL, err) // Log and exit if the URL is malformed
		return
	}

	// Extract the file name from the URL path (e.g., "myfile.pdf")
	fileName := path.Base(parsedURL.Path)
	if fileName == "" || fileName == "/" {
		log.Printf("Could not determine file name from %q", finalURL) // Log and exit if name is missing
		return
	}

	// Make sure the file has a .pdf extension
	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
		fileName += ".pdf"
	}

	// Ensure the output directory exists (create it if it doesn't)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Printf("Failed to create directory %s: %v", outputDir, err) // Log directory creation error
		return
	}

	// Build the full path to where the file will be saved
	filePath := filepath.Join(outputDir, fileName)

	// ✅ If the file already exists, skip the download
	if fileExists(filePath) {
		log.Printf("File already exists, skipping: %s", filePath)
		return
	}

	// Send the HTTP GET request to download the PDF file
	resp, err := http.Get(finalURL)
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err) // Log fetch failure
		return
	}
	defer resp.Body.Close() // Close the response body after function completes

	// Check for successful HTTP response (status 200)
	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed for %s: %s", finalURL, resp.Status) // Log error status code
		return
	}

	// Create the destination file
	outFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filePath, err) // Log file creation error
		return
	}
	defer outFile.Close() // Close file handle after writing

	// Copy the contents from the response body to the local file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Printf("Failed to save PDF to %s: %v", filePath, err) // Log writing error
		return
	}

	// Log successful download
	log.Printf("Downloaded to %s\n", filePath)
}

/*
It checks if the file exists
If the file exists, it returns true
If the file does not exist, it returns false
*/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Function to read file line by line and return a slice of strings
func readFileLines(filename string) []string {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	// Create a slice to store the lines
	var lines []string

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Append each line to the slice
		lines = append(lines, scanner.Text())
	}

	// Check if there was an error during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}

	// Return the slice of lines
	return lines
}

func main() { // Main entry point
	// List of test URLs
	urls := readFileLines("extracted_urls.txt") // Read URLs from a file

	// Path to the directory where the PDFs will be saved
	pdfDir := "./NYPD_PDF/"

	// Download counter
	downloadCount := 0
	maxDownloads := 5000

	// Loop through each input URL and convert it
	for _, url := range urls {
		if downloadCount >= maxDownloads {
			log.Printf("Reached maximum download limit of %d. Stopping.\n", maxDownloads)
			break
		}

		finalURL := extractFinalDocumentCloudURL(url)
		if finalURL == "" {
			log.Printf("Invalid or unrecognized DocumentCloud URL: %s", url)
			continue
		}

		// Only count it as a download if the file doesn't already exist
		parsedURL, err := urlParseSafe(finalURL)
		if err != nil {
			log.Printf("Skipping invalid final URL: %s", finalURL)
			continue
		}

		fileName := path.Base(parsedURL.Path)
		if fileName == "" {
			continue
		}

		filePath := filepath.Join(pdfDir, fileName)
		if !fileExists(filePath) {
			downloadPDF(finalURL, pdfDir)
			downloadCount++
		} else {
			log.Printf("File already exists, not counting as a download: %s", filePath)
		}
	}
}

// Helper function to safely parse URLs
func urlParseSafe(raw string) (*url.URL, error) {
	return url.Parse(raw)
}
