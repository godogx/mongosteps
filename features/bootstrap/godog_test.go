package bootstrap

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.nhat.io/testcontainers-go-extra"
	mongodb "go.nhat.io/testcontainers-go-registry/mongo"

	"github.com/godogx/mongosteps"
)

// Used by init().
//
//nolint:gochecknoglobals
var (
	runGoDogTests bool

	out = new(bytes.Buffer)
	opt = godog.Options{
		Strict: true,
		Output: out,
	}
)

// This has to run on init to define -godog flag, otherwise "undefined flag" error happens.
//
//nolint:gochecknoinits
func init() {
	flag.BoolVar(&runGoDogTests, "godog", false, "Set this flag is you want to run godog BDD tests")
	godog.BindCommandLineFlags("", &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()

	if runGoDogTests {
		logger := log.New(os.Stderr, "", log.Lmicroseconds|log.Lshortfile)
		mongoVersion := getEnv("MONGO_VERSION", "4.4")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		logger.Printf("starting mongo:%s container\n", mongoVersion)

		_, err := mongodb.StartGenericContainer(ctx,
			testcontainers.WithNamePrefix(randomString(8)),
			testcontainers.WithImageTag(mongoVersion),
		)
		if err != nil {
			logger.Fatalf("failed to start mongo container: %s\n", err.Error())
		}

		if opt.Randomize == 0 {
			opt.Randomize = rand.Int63() // nolint: gosec
		}
	}

	// Run the tests
	ret := m.Run()

	// Exit
	os.Exit(ret) // nolint: gocritic
}

func TestIntegration(t *testing.T) {
	if !runGoDogTests {
		t.Skip(`Missing "-godog" flag, skipping integration test.`)
	}

	deadline, ok := t.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(os.ExpandEnv(mongodb.DSN(""))))
	if err != nil {
		t.Fatalf("failed to connect to mongo: %s\n", err.Error())
	}

	m := mongosteps.NewManager(
		mongosteps.WithDefaultDatabase(conn.Database("default_db"),
			mongosteps.CleanUpAfterScenario("customer"),
		),
		mongosteps.WithDatabase("other", conn.Database("other_db"),
			mongosteps.CleanUpAfterScenario("customer"),
		),
	)

	runSuite(t, "../", func(_ *testing.T, sc *godog.ScenarioContext) {
		m.RegisterContext(sc)
	})
}

func getEnv(name, defaultValue string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultValue
	}

	return val
}

func randomString(length int) string {
	var rngSeed int64

	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed) // nolint: errcheck
	r := rand.New(rand.NewSource(rngSeed))                       // nolint: gosec

	result := make([]byte, length/2)

	_, _ = r.Read(result)

	return hex.EncodeToString(result)
}

func runSuite(t *testing.T, path string, initScenario func(t *testing.T, sc *godog.ScenarioContext)) {
	t.Helper()

	var paths []string

	files, err := ioutil.ReadDir(filepath.Clean(path))
	assert.NoError(t, err)

	paths = make([]string, 0, len(files))

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".feature") {
			paths = append(paths, filepath.Join(path, f.Name()))
		}
	}

	for _, path := range paths {
		path := path

		t.Run(path, func(t *testing.T) {
			opt.Paths = []string{path}
			suite := godog.TestSuite{
				Name:                 "Integration",
				TestSuiteInitializer: nil,
				ScenarioInitializer: func(s *godog.ScenarioContext) {
					initScenario(t, s)
				},
				Options: &opt,
			}
			status := suite.Run()

			if status != 0 {
				fmt.Println(out.String())
				assert.Fail(t, "one or more scenarios failed in feature: "+path)
			}
		})
	}
}
