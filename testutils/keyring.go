package testutils

import "embed"

//go:embed keyring-test/*
var MockKeyring embed.FS
