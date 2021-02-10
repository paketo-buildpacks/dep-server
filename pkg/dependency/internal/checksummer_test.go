package internal_test

import (
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestChecksummer(t *testing.T) {
	spec.Run(t, "checksummer", testChecksummer, spec.Report(report.Terminal{}))
}

func testChecksummer(t *testing.T, when spec.G, it spec.S) {
	var (
		assert       = assert.New(t)
		require      = require.New(t)
		checksummer  internal.Checksummer
		testDir      string
		filePath     string
		fileContents = "some-contents"
		fileMD5      = "0b9791ad102b5f5f06ef68cef2aae26e"
		fileSHA1     = "21202296bf50267250155e46d3b9eb3e4c1acb7e"
		fileSHA256   = "6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800"
		fileSHA512   = "b7b2b9e0a4d7f84985a720d1273166bb00132a60ac45388a7d3090a7d4c9692f38d019f807a02750f810f52c623362f977040231c2bbf5947170fe83686cfd9d"
		pgpKey       = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBF5r2PgBCADYbxppn7rxegXkHW8+r+I6EE0JRAcu86cjOTTj3uF8/iz/loEa
kD69l8GThUfLs7SuGaHotXWWlQg0u+xrpktBTwP+W+JUyzUNwlYFYfk4cWQjs0mn
WTbc+be190+zuELh6ywRJZ77G1cLPR0CeaM7ua/rKU9yy43eG7UK+RBgdwINs75R
SvydNkkr70Y5R2QPNSXUcYt0hls4n4XgDDM/+lSD0kc/PlNNW6SNtAi9EncmOFaZ
cIE7oltaFrmTmDMf+gzlMwDlIKPUI7rp7paqYusxmBIbFoQf8HFw+cNTP58Eynfo
9DO3osFgmxM9QOc7CEHEzbmEoMLGsRpYWnoFABEBAAG0MEJ1aWxkcGFja3MgUmVs
ZW5nIDxidWlsZHBhY2tzLXJlbGVuZ0BwaXZvdGFsLmlvPokBVAQTAQgAPhYhBHZ9
vAmtZDtjdQtuJG1ZpV7vyjjNBQJea9j4AhsDBQkDwmcABQsJCAcCBhUKCQgLAgQW
AgMBAh4BAheAAAoJEG1ZpV7vyjjNx/IIALUymfECDHxc7TtWOflrqywBSE0kf9+O
Vnq2epFLa24wxL3aVP4B1zMP0xYqA1jhC6TT5vz5c08IDVUbX9o5y/q6ZkmiTqJF
AaS55G37v6hnSz/BnN23opT8WNNanIAAydXW/WnkM7uNkjEm40qjIOBPYYMzdmsk
dvFu6mb93WW2hVFuYhfc6j3a36ditASmurPGWunDzzSUOtS+7sYubr2+RUqQlXd4
rucByXtPGiw+6verfVnTpcLCgNVUTfbZn7I7lQ3/bOiWVlRw67GC90sD98TkC55m
uKlj29Z/PE/VXuMs4ZYyBisaOv5X2uC0335/hZ757BE/eboXwpSlRVu5AQ0EXmvY
+AEIAONY+b6yYLrjYCrxRo4qY9qmcVNF+0NW4sghleN/36eJTwqnmsm+LnDsO+d0
T/Ao5dqOLBKBuqhRRI1hTQ+qKwYoJrnZHWYYMysJKw8s54EuHm3kMFQY1R/CC+lp
Vp3iGxmuw2gx64CcF/qg3TDxi2oLngLnusdYcsd+XfUFQPo4XX3ridpn0vBlJGIS
uo94rHJtvf9eG8I+x2XqZmo1bZPQIlg8o5RtUvbXyQhj8gkQJm7XHM0cXKXO4krY
FUbKMsF0gbtF+pxKf7sxhpqNfeAqvDWzNurEFEy41TOhd4GF2DoyTrELVJxY6P1D
hGEHR5WnwDIAfPCgdMHFHerVTUEAEQEAAYkBPAQYAQgAJhYhBHZ9vAmtZDtjdQtu
JG1ZpV7vyjjNBQJea9j4AhsMBQkDwmcAAAoJEG1ZpV7vyjjNqcAIAMCJ5xB5leQn
0lH1sXUOtMhNH2gVcUJOMbQsfNVNn5LvUXkcy/VNV14EpeyjjITLkwHq2onqcMd/
zkVGyiejQqOuF42a9iAi+T4Bsgyk/04EWPrglWvvfTaz0wjfsK5G0Nw1UDNR7rZX
1o4HJ97jWL1G+sjVoj/s2TRy7cl0QT1l6PQ9l5yaieGvVc24A1i5kdb+/LNJRF2V
jzDgK3Z4o3019N9gijWEvRq14jVhiPJ2eaqHjHF7mYejkZlp2e58XHNGEBlZKw22
Ho6u9LfNUwc2PYrleHakgzqXKAfEM4YadcPx4wleV7oVi3VTORyUPa+N3Y92ysZk
J5a7POjvCtQ=
=61mz
-----END PGP PUBLIC KEY BLOCK-----`
		wrongPGPKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

xo0EXv328gEEALi20eXCLaK7l4mP37FrOFE6HZtKBOMYDoK6jBrK3JQWVwEHOvba
Uy8SQZhvO/lmuNiEApVK/EK/PHsK+n1EBd45ubDic5kJYAOYj3o7KzEp0AoKm019
idlyvf2ithlwgEYuiNs8EJ1RW6sN4aHBRDoWJTy2USWd4ybrCIKP4X3LABEBAAHN
H3dyb25nLWtleSA8d3Jvbmcta2V5QHdyb25nLWtleT7CrQQTAQoAFwUCXv328gIb
LwMLCQcDFQoIAh4BAheAAAoJENvfuHlH2BGT4x0D/2tfCuyz/0I8rQdAA7yuV79T
6DYT5LF3mCM0sdb3RJ860qOm/uuKRHZVD8guZLFzb+UfY3D9K3SiDjnPfVje3LF6
WuMKiOpXgJrQoIjr/4ct8D6RdI34/gBiBBNT2U6vAYCrrkHre3iNpxU3m7/sWcA0
qDJDFr91MQH3t7giLxLnzo0EXv328gEEAMAzp33xERfo0VWDvmmnWBpNijRxX+wa
q7ChrIHr55PxArqHdrKsWXWAoNalca3a85EnJV56q1qHLe4xdqpdGOpIeuGjGvmJ
xLYAvHqe8ahjmcXzRbS4GkZLXfqCrQCjAvpqBkJ5DuqdHX9jYSDC7OzTzYhxVxHe
ka5W0pECnszBABEBAAHCwIMEGAEKAA8FAl799vIFCQ8JnAACGy4AqAkQ29+4eUfY
EZOdIAQZAQoABgUCXv328gAKCRCIoBHXMYUel8dDA/9rsfXweB15TpVaBZCb6wGN
ORtoWvzmyiHMhv94us13aTIxRq9DJeucdycaVBkjG0XrJ0BCxyoHu/R80bosqZ/J
aFZncxEjrsHz/RbToc60CmUzo7uOH4BQq924hTjdn7ReubU01fEJ5kH5y3x1weln
4xE4WWpNHapYsGpy/yIedZdZBACVyMnG0HPgvOsUwojWFGS9QrsLbKTS+39YXx9/
+E68tIOggeyNp/02nOIEBlFZmZlCmAtEEjHGrttS7bEuhugf11weHZDKdWNjgEqv
QO1EK5tMoY89m1frE+1DRYcEdMkkLiLnFPH/wx2vlFsL36vDiK94t56p34qaGPIQ
CGHFw86NBF799vIBBACp+2gStzOkiwwFHONtOWoiuMDkj23duN6TV0Rh5gRl7Lxq
25ZKgGyi3m6nJ9bIqHj69CZFqw5i6kG4uuA+IhduBjpZVmVQxB1qyrVZ3Upa9ir3
n2Iyg0z7Ie0slquXzyrgIlJ1U8u1yigyiKr7WIHJfR5FW2efPk0F46LBrzLfZQAR
AQABwsCDBBgBCgAPBQJe/fbyBQkPCZwAAhsuAKgJENvfuHlH2BGTnSAEGQEKAAYF
Al799vIACgkQdel7U6o9nntNPQQAneZcf7cC1PKKEGt1pDgmt2/j5yXkn6JPvk0V
kHNqcPrzAmqSmdrrQl5rumzBxWmRk3CJCb363jWkudHL1qzBXAPZx0ysQUchsEls
AIq1mOvHnaZeCAUbYIFZyDmmtXW1e6NJSCSJTO9SR79QIXjBDS73h81qr8mh+4bK
yc9R4ndegwP8C+zruaQLhKemsy7RozOg9Fg+jKV/53+MH9yJ6YjtFKSNEVGJWRLd
0CPiU0WpmyRqFCtJy2VX/xW866zfu9FHO9n4sG41LPqjVv8JyfIDHirh1JreoiVI
gG4nLw6FXZzsdg1EnJzhjrgU+H4k0u0Gox1svpgetqJcyTRsPsuruuE=
=JHqa
-----END PGP PUBLIC KEY BLOCK-----
`
		fileASC = `-----BEGIN PGP SIGNATURE-----

iQEzBAABCAAdFiEEdn28Ca1kO2N1C24kbVmlXu/KOM0FAl5r3aUACgkQbVmlXu/K
OM2XsAf7BXfYaOgKjB3cvOsGbeaRGpJi9TD5bnMB3k6LimsODICuvEV6JftWZt+9
YkgnmfntCGnY9O+wULVTpOXcpHxgxGPNQ40JeR6x03Qod1da/0aS6pRJYgD5cGbi
ii5ed6wkDp3wmdk1MS4Y2QJy821XYRavAafI3mw5NDwShow3V4smgqDpBGUwz0tm
T9BHrj9ruPeaHGlz5lBDKaxnpZdozxXjo2KtvJhQVslNT1RFs+LmfNJjEVvTdTgk
wAwXrUhW9rvNsVWOnlIXXwBkvtFaSVjbG5Rcm8wMZp5FTnutP8eW7+HCTf+vNPDj
KbtkeGcVHGGJtT8krUQB6f3VPT2vQg==
=+fUZ
-----END PGP SIGNATURE-----`
	)

	it.Before(func() {
		var err error
		testDir, err = ioutil.TempDir("", "external-dependency-resource-checksummer")
		require.NoError(err)

		filePath = filepath.Join(testDir, "some-file")
		err = ioutil.WriteFile(filePath, []byte(fileContents), 0644)
		require.NoError(err)

		checksummer = internal.Checksummer{}
	})

	it.After(func() {
		_ = os.RemoveAll(testDir)
	})

	when("VerifyASC", func() {
		it("verifies a file is signed with the given pgp key and signature", func() {
			err := checksummer.VerifyASC(fileASC, filePath, pgpKey)
			require.NoError(err)
		})

		when("there are multiple pgp keys", func() {
			it("succeeds if any match", func() {
				err := checksummer.VerifyASC(fileASC, filePath, wrongPGPKey, pgpKey, wrongPGPKey)
				require.NoError(err)
			})
		})

		when("the pgp key does not match", func() {
			it("returns an error", func() {
				err := checksummer.VerifyASC(fileASC, filePath, "some-bad-pgp-key")
				assert.Error(err)
			})
		})

		when("the signature does not match", func() {
			it("returns an error", func() {
				err := checksummer.VerifyASC("some-bad-asc", filePath, pgpKey)
				assert.Error(err)
			})
		})
	})

	when("VerifyMD5", func() {
		it("verifies a file's MD5", func() {
			err := checksummer.VerifyMD5(filePath, fileMD5)
			require.NoError(err)
		})

		when("the MD5 does not match", func() {
			it("returns an error", func() {
				err := checksummer.VerifyMD5(filePath, "some-bad-md5")
				assert.Error(err)
				assert.Equal("expected MD5 'some-bad-md5' but got '0b9791ad102b5f5f06ef68cef2aae26e'", err.Error())
			})
		})
	})

	when("VerifySHA1", func() {
		it("verifies a file's SHA1", func() {
			err := checksummer.VerifySHA1(filePath, fileSHA1)
			require.NoError(err)
		})

		when("the SHA does not match", func() {
			it("returns an error", func() {
				err := checksummer.VerifySHA1(filePath, "some-bad-sha")
				assert.Error(err)
				assert.Equal("expected SHA 'some-bad-sha' but got '21202296bf50267250155e46d3b9eb3e4c1acb7e'", err.Error())
			})
		})
	})

	when("VerifySHA256", func() {
		it("verifies a file's SHA256", func() {
			err := checksummer.VerifySHA256(filePath, fileSHA256)
			require.NoError(err)
		})

		when("the SHA does not match", func() {
			it("returns an error", func() {
				err := checksummer.VerifySHA256(filePath, "some-bad-sha")
				assert.Error(err)
				assert.Equal("expected SHA 'some-bad-sha' but got '6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800'", err.Error())
			})
		})
	})

	when("VerifySHA512", func() {
		it("verifies a file's SHA512", func() {
			err := checksummer.VerifySHA512(filePath, fileSHA512)
			require.NoError(err)
		})

		when("the SHA does not match", func() {
			it("returns an error", func() {
				err := checksummer.VerifySHA512(filePath, "some-bad-sha")
				assert.Error(err)
				assert.Equal("expected SHA 'some-bad-sha' but got 'b7b2b9e0a4d7f84985a720d1273166bb00132a60ac45388a7d3090a7d4c9692f38d019f807a02750f810f52c623362f977040231c2bbf5947170fe83686cfd9d'", err.Error())
			})
		})
	})

	when("GetSHA256", func() {
		it("returns a file's SHA256", func() {
			sha, err := checksummer.GetSHA256(filePath)
			require.NoError(err)
			assert.Equal(fileSHA256, sha)
		})
	})

	when("SplitPGPKeys", func() {
		it("splits a block of multiple keys into a slice of individual keys", func() {
			block := `
some garbage
-----BEGIN PGP PUBLIC KEY BLOCK-----
aaaaaaaaaa
aaaaaaaaaa
-----END PGP PUBLIC KEY BLOCK-----
more stuff

-----BEGIN PGP PUBLIC KEY BLOCK-----

bbbbbbbbbb
bbbbb
-----END PGP PUBLIC KEY BLOCK-----
-----BEGIN PGP PUBLIC KEY BLOCK-----
cccccccccc
-----END PGP PUBLIC KEY BLOCK-----
even more garbage
`

			keys := checksummer.SplitPGPKeys(block)
			assert.Equal([]string{
				`-----BEGIN PGP PUBLIC KEY BLOCK-----
aaaaaaaaaa
aaaaaaaaaa
-----END PGP PUBLIC KEY BLOCK-----`,
				`-----BEGIN PGP PUBLIC KEY BLOCK-----

bbbbbbbbbb
bbbbb
-----END PGP PUBLIC KEY BLOCK-----`,
				`-----BEGIN PGP PUBLIC KEY BLOCK-----
cccccccccc
-----END PGP PUBLIC KEY BLOCK-----`,
			}, keys)
		})
	})
}
