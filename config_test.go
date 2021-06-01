package httpbase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPCertConfiguration(t *testing.T) {
	// Non-TLS
	c := &Config{}
	assert.Nil(t, c.PostLoad())

	// Successful TLS Config
	c = &Config{HTTPCertFile: "cert", HTTPKeyFile: "key"}
	assert.Nil(t, c.PostLoad())

	// Incomplete
	c = &Config{HTTPCertFile: "cert"}
	assert.Equal(t, ErrBadCertConfig, c.PostLoad())

	c = &Config{HTTPKeyFile: "key"}
	assert.Equal(t, ErrBadCertConfig, c.PostLoad())
}
