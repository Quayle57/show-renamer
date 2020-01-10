package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/akamensky/argparse"
)

func isAVideoFile(extension string) bool {
	switch extension {
	case
		".mkv",
		".avi",
		".wmv",
		".mp4",
		".flv",
		".mpg",
		".mpeg",
		".mov",
		".m4v":
		return true
	}
	return false
}

func removeCharacters(input string, characters string) string {
	filter := func(r rune) rune {
	if strings.IndexRune(characters, r) < 0 {
		return r
	}
		return -1
	}
	return strings.Map(filter, input)
}


// readDir calls an ls on the parameter path and call the functions for each file
func readDir(directoryPath string) {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while reading path :", err)
		os.Exit(1)
	}
	// For every file, we parse the name, call the API and rename it
	var formerName, title string
	var id int
	for i, f := range files {
		fmt.Printf("Scanning file %d of %d\n", i+1, len(files))
		if !isAVideoFile(filepath.Ext(f.Name())) {
			fmt.Println(f.Name(), "is not a video file, skipping.")
		} else {
			name, season, number := parseName(f)
			if name == "" || season == "" || number == "" {
				fmt.Fprintf(os.Stderr, "Can't find the name, season or the episode number. Skipping file.\n")
				continue
			}
			if name != formerName {
				// check for new show name and id only when the filename is different
				id, title = getShowID(name)
				if id == 0 || title == "" {
					fmt.Fprintf(os.Stderr, "Cannot find episode ID or Title, Skipping file.")
					continue
				}
				formerName = name
			}
			episode := getDbInfo(id, season, number)
			if episode == (Episode{}) {
				fmt.Fprintf(os.Stderr, "Can't find episode on the API. Skipping file.\n")
				continue
			}
			formated := fmt.Sprintf("%s - S%02dE%02d - %s%s", title, episode.Season, episode.Number, removeCharacters(episode.Name, "/<>:\"\\|?*"), filepath.Ext(f.Name()))
			fmt.Println("The old name was:    ", f.Name())
			fmt.Println("The new name will be:", formated)
			filepath.Dir(directoryPath)
			oldpath := filepath.Join(directoryPath, f.Name())
			newpath := filepath.Join(directoryPath, formated)
			if oldpath == newpath {
				fmt.Println("Name is already well formated, skipping.")
			} else {
				if *autorename && !(*test) {
					err = os.Rename(oldpath, newpath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "An error occured while renaming file : %s.\n", err)
					}
				} else if *test {
					fmt.Println("Test mode, doing nothing.")
				} else {
					confirmRename(oldpath, newpath)
				}
			}
		}
		fmt.Printf("\n")
	}
	fmt.Println("No more file to check, exiting.")
}

func confirmRename(oldpath, newpath string) {
	reader := bufio.NewReader(os.Stdin)
again:
	fmt.Println("Are you sure ? (Yes/No)")
	confirm, err := reader.ReadString('\n')
	confirm = strings.TrimRight(confirm, "\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		goto again
	}
	switch strings.ToLower(confirm) {
	case
		"y",
		"yes":
		err = os.Rename(oldpath, newpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occured while renaming file : %s.\n", err)
		}
	case
		"n",
		"no":
		fmt.Println("Cancelling renaming. Going to next file.")
	default:
		fmt.Fprintf(os.Stderr, "Invalid answer : %s.\n", confirm)
		goto again
	}
}

// parseName parses the file name and return the season and the number
func parseName(file os.FileInfo) (string, string, string) {
	var re = regexp.MustCompile(*regex)
	groupNames := re.SubexpNames()
	var name, season, number string

	// parsing du nom
	var replacer = strings.NewReplacer(".", " ", "-", " ")
	match := re.FindAllStringSubmatch(file.Name(), -1)
	if len(match) > 0 {
		for groupIdx, group := range match[0] {
			switch groupNames[groupIdx] {
			case "name":
				name = strings.TrimSpace(replacer.Replace(group))
			case "season":
				season = group
			case "episode":
				number = group
			default:
			}
		}
	}
	return name, season, number
}

// getShowID calls the API with the show name to get its API ID and the API name
func getShowID(name string) (int, string) {
	url := fmt.Sprintf("http://api.tvmaze.com/search/shows?q=%s/", name)
	// get the show name and id
	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "The HTTP request failed with error %s\n", err)
		return 0, ""
	}
	// getting data in bytes
	data, _ := ioutil.ReadAll(response.Body)
	// make a dynamic array of JResponse to stock the JSON
	f := make([]JResponse, 1)
	// unserialize the JSON into the struct
	err = json.Unmarshal(data, &f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while parsing data")
		return 0, ""
	}
	return f[0].Show.ID, f[0].Show.Name
}

// getDbInfo generate the episode link for the API and return the parsed informations
func getDbInfo(id int, season, number string) Episode {
	url := fmt.Sprintf("http://api.tvmaze.com/shows/%d/episodebynumber?season=%s&number=%s/", id, season, number)
	// get the show episode infos
	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "The HTTP request failed with error %s\n", err)
		return Episode{}
	}
	data, _ := ioutil.ReadAll(response.Body)
	var episode Episode
	err = json.Unmarshal(data, &episode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while parsing data : %s", err)
		return Episode{}
	}
	return episode
}

// autorename is a parameter to bypass the confirmation for each episode
// test is a parameter that allows to run the program without renaming anything
var autorename, test *bool

// regex is the regular expression used to parse the filename
var regex *string

// main parse the arguments and launch the program
func main() {
	parser := argparse.NewParser("renamer", "Rename your show with format \"Show Name - SxxExx - Episode Name\"")
	path := parser.String("p", "path", &argparse.Options{Required: true, Help: "Path to folder to scan."})
	autorename = parser.Flag("a", "auto", &argparse.Options{Required: false, Help: "Automatically rename your show if set."})
	test = parser.Flag("t", "test", &argparse.Options{Required: false, Help: "Do a test run without renaming anything."})
	regex = parser.String("r", "regexp", &argparse.Options{Required: false, Help: "/!\\ EXPERIMENTAL /!\\ Replace the current regexp, it ABSOLUTELY needs the following capture groups : name, season and episode."})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	} else {
		if *regex == "" {
			*regex = `(?i)(\[\w+\])*((?P<name>[\S .]+)*)[. ][S ]*(?P<season>\d{2})[Ex](?P<episode>\d{1,2})`
		}
		readDir(*path)
	}
}
