package main

// findProfile returns a pointer to a profile by its name.
func findProfile(cfg *Config, name string) *Profile {
	for i, p := range cfg.Profiles {
		if p.Name == name {
			return &cfg.Profiles[i]
		}
	}
	return nil
}