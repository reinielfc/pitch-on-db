// Package services implements the business logic layer of the application.
// Each service depends on one or more [repos] interfaces, which are injected
// via constructor functions, keeping the business logic decoupled from the
// underlying data store and testable with mock implementations.
//
// The main service is [PigeonService], which orchestrates pigeon record
// management, genealogy (parent/child assignments), and tag operations.
package services
