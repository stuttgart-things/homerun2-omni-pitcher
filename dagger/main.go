// A generated module for Dagger functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/dagger/internal/dagger"
	"fmt"
)

type Dagger struct{}

func (m *Dagger) BuildAndTestBinary(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default="1.25.4"
	goVersion string,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="amd64"
	arch string,
	// +optional
	// +default="main.go"
	goMainFile string,
	// +optional
	// +default="main"
	binName string,
	// +optional
	// +default=""
	ldflags string,
	// +optional
	// +default="."
	packageName string,
	// +optional
	// +default="./..."
	testPath string,
	// +optional
	// +default="8080"
	port int,
) (*dagger.File, error) {

	// Build the binary
	binDir := dag.Go().BuildBinary(
		source,
		dagger.GoBuildBinaryOpts{
			GoVersion:   goVersion,
			Os:          os,
			Arch:        arch,
			GoMainFile:  goMainFile,
			BinName:     binName,
			Ldflags:     ldflags,
			PackageName: packageName,
		})

	// Start Redis service for testing
	redisService := dag.Homerun().RedisService(dagger.HomerunRedisServiceOpts{
		Version:  "7.2.0-v18",
		Password: "",
	})

	// Build test command: start binary, test health endpoint, test pitch endpoint, stop binary
	testCmd := fmt.Sprintf(`
exec > /app/test-output.log 2>&1
set -e

echo "=== Starting binary ==="
./%s &
BIN_PID=$!
sleep 3

echo ""
echo "=== Testing health endpoint ==="
curl -f -v http://localhost:%d/health || {
  echo "Health check failed!"
  kill $BIN_PID 2>/dev/null || true
  exit 1
}

echo ""
echo "=== Testing pitch endpoint ==="
curl -f -v -X POST http://localhost:%d/pitch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token-12345" \
  -d '{
    "title": "Test Notification",
    "message": "Testing Redis integration",
    "severity": "info",
    "author": "dagger-test",
    "system": "test-system",
    "tags": "test",
    "assigneeaddress": "test@example.com",
    "assigneename": "Test User"
  }' || {
  echo "Pitch endpoint failed!"
  kill $BIN_PID 2>/dev/null || true
  exit 1
}

echo ""
echo "=== All tests passed! ==="
kill $BIN_PID 2>/dev/null || true
exit 0
`, binName, port, port)

	result := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "curl"}, dagger.ContainerWithExecOpts{}).
		WithDirectory("/app", binDir).
		WithWorkdir("/app").
		WithServiceBinding("redis", redisService).
		WithEnvVariable("REDIS_ADDR", "redis").
		WithEnvVariable("REDIS_PORT", "6379").
		WithEnvVariable("REDIS_STREAM", "messages").
		WithEnvVariable("AUTH_TOKEN", "test-token-12345").
		WithExec([]string{"sh", "-c", testCmd}, dagger.ContainerWithExecOpts{})

	// Check if tests passed
	_, err := result.Sync(ctx)
	if err != nil {
		// Return logs even on failure
		testLog := result.File("/app/test-output.log")
		return testLog, fmt.Errorf("tests failed - check test-output.log for details: %w", err)
	}

	// Return the test log file
	testLog := result.File("/app/test-output.log")
	return testLog, nil
}

// func (m *Dagger) BuildAndTestBinary(
// 	ctx context.Context,
// 	source *dagger.Directory,
// 	// +optional
// 	// +default="1.25.4"
// 	goVersion string,
// 	buildPath string,
// 	testPath string,
// 	// +optional
// 	// +default="linux"
// 	os string,
// 	// +optional
// 	// +default="amd64"
// 	arch string,
// 	// +optional
// 	// +default="main"
// 	binName string,
// 	// +optional
// 	// +default=""
// 	ldflags string,
// 	// +optional
// 	// +default="GITHUB_TOKEN"
// 	tokenName string,
// 	// +optional
// 	// +default=""
// 	packageName string,
// 	// +optional
// 	token *dagger.Secret,
// 	// +optional
// 	// +default="ko.local"
// 	koRepo string,
// 	// +optional
// 	// +default="v0.18.0"
// 	koVersion string,
// 	// +optional
// 	// +default="."
// 	koBuildArg string,
// 	// +optional
// 	// +default="false"
// 	koPush string,
// 	// +optional
// 	// +default=true
// 	buildBinary bool,
// 	// +optional
// 	// +default=true
// 	koBuild bool,
// 	// +optional
// 	// +default="build-report.txt"
// 	reportName string) (string, error) {

// 	buildDir := dag.GoMicroservice().RunBuildStage(
// 		source,
// 		dagger.GoMicroserviceRunBuildStageOpts{
// 			GoVersion:   goVersion,
// 			Os:          os,
// 			Arch:        arch,
// 			GoMainFile:  buildPath,
// 			BinName:     binName,
// 			Ldflags:     ldflags,
// 			TokenName:   tokenName,
// 			PackageName: packageName,
// 			Token:       token,
// 			KoRepo:      koRepo,
// 			KoVersion:   koVersion,
// 			KoBuildArg:  koBuildArg,
// 			KoPush:      koPush,
// 			BuildBinary: buildBinary,
// 			KoBuild:     koBuild,
// 			ReportName:  reportName,
// 		},
// 	)

// 	return buildDir.File(reportName).Contents(ctx)
// }

// func (m *Dagger) RunGoTests(
// 	ctx context.Context,
// 	source *dagger.Directory,
// 	// +optional
// 	// +default="1.25.4"
// 	goVersion string,
// 	// +optional
// 	// +default="7.2.0-v18"
// 	redisVersion string,
// 	testPath string,
// ) (string, error) {

// 	return dag.
// 		Homerun().
// 		RunTestWithRedis(
// 			ctx,
// 			source,
// 			testPath,
// 			dagger.HomerunRunTestWithRedisOpts{
// 				GoVersion:    goVersion,
// 				RedisVersion: redisVersion,
// 			},
// 		)
// }
