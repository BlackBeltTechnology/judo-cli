package db

import (
	"fmt"
	"os"
	"path/filepath"

	"judo-cli-module/internal/utils"
)

// DumpPostgresql dumps the PostgreSQL database to a file.
func DumpPostgresql(containerName, schema string) (string, error) {
	timestamp := utils.TimeNow().Format("20060102_150405")
	file := fmt.Sprintf("%s_dump_%s.tar.gz", schema, timestamp)

	out, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Use pg_dump custom format (-F c), same as bash
	cmd := utils.ExecuteCommand("docker", "exec", "-i", containerName,
		"/bin/bash", "-c",
		fmt.Sprintf("PGPASSWORD=%s pg_dump --username=%s -F c %s", schema, schema, schema))
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return file, cmd.Run()
}

// ImportPostgresql imports a PostgreSQL database dump.
func ImportPostgresql(containerName, schema, dumpFile string) error {
	in, err := os.Open(dumpFile)
	if err != nil {
		return err
	}
	defer in.Close()

	cmd := utils.ExecuteCommand("docker", "exec", "-i", containerName,
		"/bin/bash", "-c",
		fmt.Sprintf("PGPASSWORD=%s pg_restore -Fc --clean -U %s -d %s", schema, schema, schema))
	cmd.Stdin = in
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FindLatestDump finds the latest PostgreSQL dump file for a given schema.
func FindLatestDump(schema string) (string, error) {
	pattern := fmt.Sprintf("%s_dump_*.tar.gz", schema)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no dump files found matching %q", pattern)
	}

	// Filenames include timestamp; lexicographic max is the latest
	latest := matches[0]
	for _, m := range matches[1:] {
		if m > latest {
			latest = m
		}
	}
	return latest, nil
}