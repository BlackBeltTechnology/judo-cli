package db

import (
	"context"
	"fmt"
	"io"
	"judo-cli-module/internal/utils"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var cli *client.Client

func init() {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
}

// DumpPostgresql dumps the PostgreSQL database to a file.
func DumpPostgresql(containerName, schema string) (string, error) {
	timestamp := utils.TimeNow().Format("20060102_150405")
	file := fmt.Sprintf("%s_dump_%s.tar.gz", schema, timestamp)

	out, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer out.Close()

	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"/bin/bash", "-c",
			fmt.Sprintf("PGPASSWORD=%s pg_dump --username=%s -F c %s", schema, schema, schema),
		},
	}

	resp, err := cli.ContainerExecCreate(context.Background(), containerName, execConfig)
	if err != nil {
		return "", err
	}

	hijackedResponse, err := cli.ContainerExecAttach(context.Background(), resp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", err
	}
	defer hijackedResponse.Close()

	_, err = io.Copy(out, hijackedResponse.Reader)
	return file, err
}

// ImportPostgresql imports a PostgreSQL database dump.
func ImportPostgresql(containerName, schema, dumpFile string) error {
	in, err := os.Open(dumpFile)
	if err != nil {
		return err
	}
	defer in.Close()

	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"/bin/bash", "-c",
			fmt.Sprintf("PGPASSWORD=%s pg_restore -Fc --clean -U %s -d %s", schema, schema, schema),
		},
	}

	resp, err := cli.ContainerExecCreate(context.Background(), containerName, execConfig)
	if err != nil {
		return err
	}

	hijackedResponse, err := cli.ContainerExecAttach(context.Background(), resp.ID, container.ExecStartOptions{})
	if err != nil {
		return err
	}
	defer hijackedResponse.Close()

	_, err = io.Copy(hijackedResponse.Conn, in)
	if err != nil {
		return err
	}

	return nil
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
