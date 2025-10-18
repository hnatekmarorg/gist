package gist

// FindProfile returns a pointer to a profile by its name.
func FindProfile(cfg *Config, name string) *Profile {
	for i, p := range cfg.Profiles {
		if p.Name == name {
			return &cfg.Profiles[i]
		}
	}
	return nil
}