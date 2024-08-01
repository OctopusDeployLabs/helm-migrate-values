package pkg

import "sort"

type MockMigrations struct {
	migrations       []Migration
	migrationDataMap map[string][]byte
}

func NewMockMigrations() *MockMigrations {
	return &MockMigrations{
		migrationDataMap: make(map[string][]byte),
	}
}

func (ms *MockMigrations) AddMigrationData(m *Migration, data string) {
	ms.migrationDataMap[m.From.String()+"-"+m.To.String()] = []byte(data)
}

func (ms *MockMigrations) GetSortedMigrations() ([]Migration, error) {
	sort.Slice(ms.migrations, func(i, j int) bool {
		return ms.migrations[i].From.LessThan(&ms.migrations[j].From)
	})
	return ms.migrations, nil
}

func (ms *MockMigrations) GetDataForMigration(m *Migration) ([]byte, error) {
	return ms.migrationDataMap[m.From.String()+"-"+m.To.String()], nil
}
