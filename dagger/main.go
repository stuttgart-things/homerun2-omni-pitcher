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
)

type Dagger struct{}

func (m *Dagger) RunGoTests(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default="1.25.4"
	goVersion string,
	// +optional
	// +default="7.2.0-v18"
	redisVersion string,
	testPath string,
) (string, error) {

	return dag.
		Homerun().
		RunTestWithRedis(
			ctx,
			source,
			testPath,
			dagger.HomerunRunTestWithRedisOpts{
				GoVersion:    goVersion,
				RedisVersion: redisVersion,
			},
		)
}
