package widevine

import (
	_ "embed"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iyear/gowidevine/device"
	wvpb "github.com/iyear/gowidevine/widevinepb"
)

var l3cdm device.Device

func init() {
	for _, l3 := range device.L3 {
		if l3.SystemID == 4464 {
			l3cdm = l3
			break
		}
	}
}

func TestRandomBytes(t *testing.T) {
	cdm := NewCDM()

	assert.Equal(t, 16, len(cdm.randomBytes(16)))
	assert.Equal(t, 32, len(cdm.randomBytes(32)))
}

type fakeSource struct{}

func (f fakeSource) Int63() int64 { return 0 }
func (f fakeSource) Seed(_ int64) {}

//go:embed testdata/pssh
var psshData []byte

//go:embed testdata/license-challenge
var licenseChallenge []byte

//go:embed testdata/license
var license []byte

func TestNewCDM(t *testing.T) {
	cdm := NewCDM(
		WithDevice(l3cdm),
		WithRandom(fakeSource{}),
		WithNow(func() time.Time { return time.Unix(0, 0) }),
	)
	require.NotNil(t, cdm)

	pssh, err := NewPSSH(psshData)
	require.NoError(t, err)

	cert, err := ParseServiceCert(serviceCert)
	require.NoError(t, err)

	challenge, parseLicense, err := cdm.GetLicenseChallenge(pssh, wvpb.LicenseType_AUTOMATIC, true, cert)
	require.NoError(t, err)

	require.Equal(t, licenseChallenge, challenge)

	// parse license
	keys, err := parseLicense(license)
	require.NoError(t, err)

	require.Len(t, keys, 1)
	assert.Equal(t, wvpb.License_KeyContainer_CONTENT, keys[0].Type)
	assert.Equal(t, "8421e83ef1d57ee79e4aaa4b0b38df47", hex.EncodeToString(keys[0].IV))
	assert.Equal(t, "df6ef2f5fd83078091a78566c8d01925", hex.EncodeToString(keys[0].ID))
	assert.Equal(t, "20be4041a33c7a081e43b2b4378d6d5c", hex.EncodeToString(keys[0].Key))
}
