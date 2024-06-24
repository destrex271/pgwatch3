package sources_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cybertec-postgresql/pgwatch3/sources"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var ctx = context.Background()

func TestKind_IsValid(t *testing.T) {
	tests := []struct {
		kind     sources.Kind
		expected bool
	}{
		{kind: sources.SourcePostgres, expected: true},
		{kind: sources.SourcePostgresContinuous, expected: true},
		{kind: sources.SourcePgBouncer, expected: true},
		{kind: sources.SourcePgPool, expected: true},
		{kind: sources.SourcePatroni, expected: true},
		{kind: sources.SourcePatroniContinuous, expected: true},
		{kind: sources.SourcePatroniNamespace, expected: true},
		{kind: "invalid", expected: false},
	}

	for _, tt := range tests {
		got := tt.kind.IsValid()
		assert.True(t, got == tt.expected, "IsValid(%v) = %v, want %v", tt.kind, got, tt.expected)
	}
}

func TestMonitoredDatabase_Connect(t *testing.T) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, pgContainer.Terminate(ctx)) }()

	// Create a new MonitoredDatabase instance
	md := &sources.MonitoredDatabase{}
	md.ConnStr, err = pgContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// Call the Connect method
	err = md.Connect(ctx)
	assert.NoError(t, err)

	// Check cached connection
	err = md.Connect(ctx)
	assert.NoError(t, err)
}
func TestMonitoredDatabase_GetDatabaseName(t *testing.T) {
	md := &sources.MonitoredDatabase{
		ConnStr: "postgres://user:password@localhost:5432/mydatabase",
	}
	expected := "mydatabase"
	got := md.GetDatabaseName()
	assert.Equal(t, expected, got, "GetDatabaseName() = %v, want %v", got, expected)

	md.ConnStr = "foo boo"
	expected = ""
	got = md.GetDatabaseName()
	assert.Equal(t, expected, got, "GetDatabaseName() = %v, want %v", got, expected)
}
func TestMonitoredDatabase_IsPostgresSource(t *testing.T) {
	md := &sources.MonitoredDatabase{
		Kind: sources.SourcePostgres,
	}
	assert.True(t, md.IsPostgresSource(), "IsPostgresSource() = false, want true")

	md.Kind = sources.SourcePgBouncer
	assert.False(t, md.IsPostgresSource(), "IsPostgresSource() = true, want false")

	md.Kind = sources.SourcePgPool
	assert.False(t, md.IsPostgresSource(), "IsPostgresSource() = true, want false")

	md.Kind = sources.SourcePatroni
	assert.True(t, md.IsPostgresSource(), "IsPostgresSource() = false, want true")
}

func TestMonitoredDatabase_ResolveDatabasesFromPostgres(t *testing.T) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase("mydatabase"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, pgContainer.Terminate(ctx)) }()

	// Create a new MonitoredDatabase instance
	md := &sources.MonitoredDatabase{}
	md.ConnStr, err = pgContainer.ConnectionString(ctx)
	assert.NoError(t, err)

	// Call the ResolveDatabasesFromPostgres method
	dbs, err := md.ResolveDatabasesFromPostgres()
	assert.NoError(t, err)
	assert.True(t, len(dbs) == 2) //postgres and mydatabase
}
