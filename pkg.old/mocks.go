package pkg_old

type MockMigrations struct {
	migrations       []Migration
	migrationDataMap map[string][]byte
}

func NewMockMigrations() *MockMigrations {
	return &MockMigrations{
		migrationDataMap: make(map[string][]byte),
	}
}

func (ms *MockMigrations) AddMigrationData(m Migration, data string) {
	ms.migrations = append(ms.migrations, m)
	ms.migrationDataMap[m.ToVersion.String()] = []byte(data)
}

func (ms *MockMigrations) GetMigrations() ([]Migration, error) {
	return ms.migrations, nil
}

func (ms *MockMigrations) GetDataForMigration(m *Migration) ([]byte, error) {
	return ms.migrationDataMap[m.ToVersion.String()], nil
}
