package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type config struct {
	App struct {
		Remote struct {
			SiteRepo  string `yaml:"site_repo"`
			PostsRepo string `yaml:"posts_repo"`
			Username  string `yaml:"username"`
			Password  string `yaml:"password"`
		}
	}
}

type proxy struct {
	Server http.Handler
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Server.ServeHTTP(w, r)
}

const (
	BUILD_A               = "BUILD_A"
	BUILD_B               = "BUILD_B"
	SITE_DIR              = "./jekyll"
	POSTS_DIR             = "./jekyll/_posts"
	UPLOADS_DIR           = "./jekyll/uploads"
	POSTS_REPO_DIR        = "./posts"
	DEFAULT_SETTINGS_PATH = "./defaults.yml"
	SETTINGS_PATH         = "./settings.yml"
)

var (
	settings         config
	jekyllRepo       string
	postsRepo        string
	currentBuild     string
	buildAFileServer http.Handler
	buildBFileServer http.Handler

	mu               sync.Mutex
	currentServer    proxy
	currentPostsHash string
	currentSiteHash  string
)

func main() {
	// Read our config from settings.yml
	err := readSettings()
	if err != nil {
		log.Fatal(err)
	}

	// Construct the repo strings (with username/password)
	jekyllRepo = fmt.Sprintf(
		settings.App.Remote.SiteRepo,
		settings.App.Remote.Username,
		settings.App.Remote.Password,
	)

	postsRepo = fmt.Sprintf(
		settings.App.Remote.PostsRepo,
		settings.App.Remote.Username,
		settings.App.Remote.Password,
	)

	// Make sure the build folders exist
	err = initStaticDirectories()
	if err != nil {
		log.Fatal("Could not init Build Directories")
	}

	// Get updated Jekyll Site
	err = updateRepo(jekyllRepo, SITE_DIR)
	if err != nil {
		log.Fatal("Could not update from Jekyll Repo")
	}

	// Get updated posts
	err = updateRepo(postsRepo, POSTS_REPO_DIR)
	if err != nil {
		log.Fatal("Could not update from Posts Repo")
	}

	// Track the jekyll repo's latest hash
	currentSiteHash, err = getMasterHash(SITE_DIR)
	if err != nil {
		log.Fatal("Could not get SHA1 hash of origin/master from Site Repo")
	}

	// Track the posts repo's latest hash
	currentPostsHash, err = getMasterHash(POSTS_REPO_DIR)
	if err != nil {
		log.Fatal("Could not get SHA1 hash of origin/master from Posts Repo")
	}

	// Symlink the "posts" and "uploads" folder into the jekyll folder
	err = symlinkDirs()
	if err != nil {
		log.Println("Could not symlink posts and uploads dir into jekyll dir")
		log.Fatal(err)
	}

	// Run Jekyll Command to Build Site
	err = buildSite("./" + BUILD_A)
	if err != nil {
		log.Fatal("Could not build site")
	}

	// Our http handlers
	buildAFileServer = http.FileServer(http.Dir("./" + BUILD_A))
	buildBFileServer = http.FileServer(http.Dir("./" + BUILD_B))

	// We initially build into "BUILD_A" folder
	currentBuild = BUILD_A
	currentServer.Server = buildAFileServer

	// Serve the site
	mux := http.NewServeMux()
	mux.Handle("/", &currentServer)

	// Our endpoint to handle git webhook
	mux.HandleFunc("/__updateposts__", handleGitWebhook)
	http.ListenAndServe(":8080", mux)
}

func handleGitWebhook(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Get up-to-date jekyll site + posts
	err := updateRepo(jekyllRepo, SITE_DIR)
	if err != nil {
		log.Printf("Could not update Jekyll Repo: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = updateRepo(postsRepo, POSTS_REPO_DIR)
	if err != nil {
		log.Printf("Could not update Posts Repo: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get the latest origin/master hashes
	postsHash, err := getMasterHash(POSTS_REPO_DIR)
	if err != nil {
		log.Printf("Could not get updated post hash: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	siteHash, err := getMasterHash(SITE_DIR)
	if err != nil {
		log.Printf("Could not get updated site hash: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Symlink the "posts" and "uploads" folder into the jekyll folder
	err = symlinkDirs()
	if err != nil {
		log.Printf("Could not resymlink: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Try to rebuild the site if the hashes are different
	if (postsHash != currentPostsHash) || (siteHash != currentSiteHash) {
		dest := "./" + BUILD_A
		if currentBuild == BUILD_A {
			dest = "./" + BUILD_B
		}

		err = buildSite(dest)
		if err != nil {
			log.Println("Error occured while trying to build the site")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Build was successful; we update the current hash to track
		// the new hashes and begin serving the site from the new build folder
		currentPostsHash = postsHash
		currentSiteHash = siteHash
		if currentBuild == BUILD_A {
			currentBuild = BUILD_B
			currentServer.Server = buildBFileServer
		} else {
			currentBuild = BUILD_A
			currentServer.Server = buildAFileServer
		}

		log.Printf("Successfully Built Site to: %s", currentBuild)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Read settings from Yaml Files
func readSettings() (err error) {
	defaultSettings, err := ioutil.ReadFile(DEFAULT_SETTINGS_PATH)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(defaultSettings, &settings)
	if err != nil {
		return err
	}

	// Check if there is a settings.yml to read (for overrides)
	if _, err := os.Stat(SETTINGS_PATH); err == nil {
		fileSettings, err := ioutil.ReadFile(SETTINGS_PATH)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(fileSettings, &settings)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create BUILD Folders
func initStaticDirectories() (err error) {
	if _, err := os.Stat("./" + BUILD_A); os.IsNotExist(err) {
		err = os.Mkdir("./"+BUILD_A, 0755)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	if _, err := os.Stat("./" + BUILD_B); os.IsNotExist(err) {
		err = os.Mkdir("./"+BUILD_B, 0755)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// Update any Repos
func updateRepo(repo string, dest string) (err error) {
	_, err = os.Stat(dest)
	// Clone it if the folder doesn't exist
	if err != nil {
		if os.IsNotExist(err) {
			err = exec.Command("git", "clone", repo, dest).Run()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Update otherwise with "git fetch; git reset; git clean"
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = dest
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "reset", "origin/master", "--hard")
	cmd.Dir = dest
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "clean", "-f", "-x", "-d")
	cmd.Dir = dest
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func symlinkDirs() (err error) {
	err = os.RemoveAll(SITE_DIR + "/_posts")
	if err != nil {
		return err
	}
	err = os.RemoveAll(SITE_DIR + "/uploads")
	if err != nil {
		return err
	}

	postsAbs, err := filepath.Abs(POSTS_REPO_DIR + "/posts")
	if err != nil {
		return errors.New("Unable to get absolute path of posts directory")
	}

	uploadsAbs, err := filepath.Abs(POSTS_REPO_DIR + "/uploads")
	if err != nil {
		return errors.New("Unable to get absolute path of uploads directory")
	}
	err = os.Symlink(postsAbs, SITE_DIR+"/_posts")
	if err != nil {
		return err
	}
	err = os.Symlink(uploadsAbs, SITE_DIR+"/uploads")
	if err != nil {
		return err
	}
	return nil
}

// Get the hash of origin/master for a repo
func getMasterHash(repoDir string) (hash string, err error) {
	cmd := exec.Command("git", "rev-parse", "origin/master")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Build The Site
func buildSite(dest string) (err error) {
	return exec.Command("jekyll", "build", "--source", SITE_DIR, "--destination", dest).Run()
}
