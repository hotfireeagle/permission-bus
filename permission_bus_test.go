package permissionbus

import (
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	filePath := filepath.Join(".", "example.json")
	filePath = filepath.Clean(filePath)

	pb, err := Load(filePath)

	if err != nil {
		t.Errorf(err.Error())
	}

	if pb.configData[0].Name != "权限管理" {
		t.Errorf("wrong parse")
	}
}
