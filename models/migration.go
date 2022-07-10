package models

func MigrationTargets() []interface{} {
	return []interface{}{
		&UsageStat{},
	}
}
