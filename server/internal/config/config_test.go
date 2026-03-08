package config

import "testing"

func TestFromEnv_AdminCredentialsDefaultToAdmin(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")

	cfg := FromEnv()

	if cfg.AdminUsername != "admin" {
		t.Fatalf("unexpected default admin username: got %q want %q", cfg.AdminUsername, "admin")
	}
	if cfg.AdminPassword != "admin" {
		t.Fatalf("unexpected default admin password: got %q want %q", cfg.AdminPassword, "admin")
	}
}

func TestFromEnv_AdminCredentialsCanBeOverridden(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "root")
	t.Setenv("ADMIN_PASSWORD", "s3cret")

	cfg := FromEnv()

	if cfg.AdminUsername != "root" {
		t.Fatalf("unexpected admin username override: got %q want %q", cfg.AdminUsername, "root")
	}
	if cfg.AdminPassword != "s3cret" {
		t.Fatalf("unexpected admin password override: got %q want %q", cfg.AdminPassword, "s3cret")
	}
}
