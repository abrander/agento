package core

import (
	"encoding/json"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/abrander/agento/userdb"
)

type Config struct {
	Hosts map[string]toml.Primitive `toml:"host"`
}

func TestHostDecodeTOML(t *testing.T) {
	cases := map[string]*Host{
		`[host.testhost]
        transport = "localtransport"
        name = "testhost"
        `: &Host{
			AccountID:   userdb.God.GetAccountId(),
			Name:        "testhost",
			TransportID: "localtransport",
		},
		`[host.testhost]
        transport = "sshtransport"
        name = "testhost2"
        host = "dev.google.com"
        `: &Host{
			AccountID:   userdb.God.GetAccountId(),
			Name:        "testhost2",
			TransportID: "sshtransport",
		},
		`[host.testhost]
        transport = "sshtransport"
        host = "127.0.0.1"
        port = 200
        `: &Host{
			AccountID:   userdb.God.GetAccountId(),
			TransportID: "sshtransport",
		},
	}

	for conf, correct := range cases {
		c := Config{}
		_, err := toml.Decode(conf, &c)

		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}

		for _, prim := range c.Hosts {
			host := Host{}
			err = host.DecodeTOML(prim)
			if err != nil && correct != nil {
				t.Fatalf("DecodeTOML error: %s", err.Error())
			}

			if host.TransportID != correct.TransportID {
				t.Errorf("Host transport %+v doesn't match correct transportID %+v", host.TransportID, correct.TransportID)
			}
		}
	}
}

func TestHostDecodeJSON(t *testing.T) {
	cases := map[string]*Host{
		`{
		  "_id": "8764786ab76dbc76e",
		  "name": "testhost",
		  "transport": "localtransport",
		  "accountId": "hejhejsa"
		}`: &Host{
			ID:          "8764786ab76dbc76e",
			AccountID:   "hejhejsa",
			Name:        "testhost",
			TransportID: "localtransport",
		},
		`{
		  "_id": "",
		  "name": "testhost2",
		  "transport": "sshtransport",
		  "host": "dev.google.com"
	    }`: &Host{
			Name:        "testhost2",
			TransportID: "sshtransport",
		},
		`{
		  "transport": "sshtransport",
		  "host": "127.0.0.1",
		  "port": 200
		}`: &Host{
			TransportID: "sshtransport",
		},
		`{
		  "transport": "nonexisting",
	    }`: nil,
		`invalid json`: nil,
	}

	for data, correct := range cases {
		host := Host{}
		err := json.Unmarshal([]byte(data), &host)
		if err != nil && correct != nil {
			t.Fatalf("json.Unmarshal error: %s", err.Error())
		}

		if correct == nil {
			if err == nil {
				t.Errorf("DecodeTOML didn't catch error")
			}
			break
		}

		if host.TransportID != correct.TransportID {
			t.Errorf("Host transport %+v doesn't match correct transportID %+v", host.TransportID, correct.TransportID)
		}
	}
}
