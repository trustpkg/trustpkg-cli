package main

import "testing"

func TestValidatePackageName(t *testing.T) {
	for _, name := range []string{"", "gcc", "openssl-devel", "libstdc++-devel", "libfoo.x86_64"} {
		if err := validatePackageName(name); err != nil {
			t.Errorf("valid name %q rejected: %v", name, err)
		}
	}
	for _, name := range []string{"-y", "gcc --all", "gcc;echo", "$(id)", "gcc\nother"} {
		if err := validatePackageName(name); err == nil {
			t.Errorf("unsafe name %q accepted", name)
		}
	}
}
