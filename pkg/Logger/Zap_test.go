package Logger

import (
	"go.uber.org/zap"
	"testing"
)

func TestZapLogger(t *testing.T) {
	GetLogger().Info("sada",zap.String("test","test1"))
}
