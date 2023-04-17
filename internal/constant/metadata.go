package constant

import (
	"encoding/json"
	"github.com/gofrs/uuid"
)

type NetWork int

func (n NetWork) String() string {
	if n == TCP {
		return "tcp"
	}
	return "udp"
}

func (n NetWork) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

type Type int

func (t Type) String() string {
	switch t {
	case HTTP:
		return "HTTP"
	case HTTPCONNECT:
		return "HTTPS"
	case SOCKS4:
		return "Socks4"
	case SOCKS5:
		return "Socks5"
	default:
		return "Unknown"
	}
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

type Metadata struct {
	ID      uuid.UUID `json:"-"`
	NetWork NetWork   `json:"network"`
	Type    Type      `json:"type"`
	Src     string    `json:"src"`
	Dest    string    `json:"dest"`
}
