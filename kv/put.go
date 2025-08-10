//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kv

import (
	"time"
)

// Put stores a new version of key-value pairs at the specified path in the
// store. It implements automatic versioning with a maximum of 3 versions per
// path.
//
// When storing values:
//   - If the path doesn't exist, it creates a new secret with initial metadata
//   - Each put operation creates a new version with an incremented version
//     number
//   - Old versions are automatically pruned when exceeding MaxVersions
//     (default: 10)
//   - Timestamps are updated for both creation and modification times
//
// Parameters:
//   - path: The location where the secret will be stored
//   - values: A map of key-value pairs to store at this path
//
// The function maintains metadata including:
//   - CreatedTime: When the secret was first created
//   - UpdatedTime: When the most recent version was added
//   - CurrentVersion: The latest version number
//   - OldestVersion: The oldest available version number
//   - MaxVersions: Maximum number of versions to keep (fixed at 10)
func (kv *KV) Put(path string, values map[string]string) {
	rightNow := time.Now()

	secret, exists := kv.data[path]
	if !exists {
		secret = &Value{
			Versions: make(map[int]Version),
			Metadata: Metadata{
				CreatedTime: rightNow,
				UpdatedTime: rightNow,
				MaxVersions: kv.maxSecretVersions,
				// Versions start at 1, so that passing 0 as the version will
				// default to the current version.
				CurrentVersion: 1,
				OldestVersion:  1,
			},
		}
		kv.data[path] = secret
	} else {
		secret.Metadata.CurrentVersion++
	}

	newVersion := secret.Metadata.CurrentVersion

	// Add a new version:
	secret.Versions[newVersion] = Version{
		Data:        values,
		CreatedTime: rightNow,
		Version:     newVersion,
	}

	// Update metadata
	secret.Metadata.UpdatedTime = rightNow

	// Clean up the old versions if exceeding MaxVersions
	var deletedAny bool
	for version := range secret.Versions {
		if newVersion-version >= secret.Metadata.MaxVersions {
			delete(secret.Versions, version)
			deletedAny = true
		}
	}

	if deletedAny {
		oldestVersion := secret.Metadata.CurrentVersion
		for version := range secret.Versions {
			if version < oldestVersion {
				oldestVersion = version
			}
		}
		secret.Metadata.OldestVersion = oldestVersion
	}
}
