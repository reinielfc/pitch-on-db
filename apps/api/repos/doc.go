// Package repos defines the data-access layer interfaces and their PostgreSQL
// implementations. Each repository wraps the sqlc-generated [db.Queries] and
// translates database rows into [domain] types.
//
// The two main repositories are:
//   - [PigeonRepository] – CRUD and genealogy operations for pigeon records.
//   - [TagRepository] – tag management and pigeon-tag association operations.
package repos
