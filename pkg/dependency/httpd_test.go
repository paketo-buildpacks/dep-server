package dependency_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
)

func TestHttpd(t *testing.T) {
	spec.Run(t, "httpd", testHttpd, spec.Report(report.Terminal{}))
}

func testHttpd(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		httpd                dependency.Dependency
	)

	removeLinesContaining := func(s, pattern string) string {
		re := regexp.MustCompile(fmt.Sprintf("(?m)[\r\n]+^.*%s.*$", pattern))
		return re.ReplaceAllString(s, "")
	}

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		httpd, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("httpd")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all httpd release versions with the newest first", func() {
			fakeWebClient.GetReturns([]byte(fullHTTPDIndex), nil)

			versions, err := httpd.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"2.4.43",
				"2.4.41",
				"2.4.39",
				"2.4.38",
				"2.2.14",
				"2.0.63",
				"2.2.6",
				"2.0.61",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("http://archive.apache.org/dist/httpd/?F=2&C=M&O=D&P=httpd-*.tar.bz2*", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		var expectedReleaseDate = time.Date(2020, 03, 30, 14, 21, 0, 0, time.UTC)

		it("returns the correct httpd version", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(httpdIndex2443), nil)

			fakeWebClient.GetReturnsOnCall(1, []byte("some-sha256 *httpd-2.4.43.tar.bz2"), nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/httpd@2.4.43?checksum=some-sha256&download_url=http://archive.apache.org/dist")

			actualDepVersion, err := httpd.GetDependencyVersion("2.4.43")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedDepVersion := dependency.DepVersion{
				Version:         "2.4.43",
				URI:             "http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:apache:http_server:2.4.43:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/httpd@2.4.43?checksum=some-sha256&download_url=http://archive.apache.org/dist",
				Licenses:        []string{"MIT", "MIT-2"},
			}

			assert.Equal(expectedDepVersion, actualDepVersion)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("http://archive.apache.org/dist/httpd/?F=2&C=M&O=D&P=httpd-2.4.43.tar.bz2*", urlArg)

			urlArg, _ = fakeWebClient.GetArgsForCall(1)
			assert.Equal("http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2.sha256", urlArg)
		})

		when("there is no SHA256 file", func() {
			it("returns the correct httpd version using the SHA1 file", func() {
				index := removeLinesContaining(httpdIndex2443, "httpd-2.4.43.tar.bz2.sha256")
				fakeWebClient.GetReturnsOnCall(0, []byte(index), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte("some-sha1 *httpd-2.4.43.tar.bz2"), nil)

				fakeChecksummer.GetSHA256Returns("some-sha256", nil)

				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/httpd@2.4.43?checksum=some-sha256&download_url=http://archive.apache.org/dist")

				actualDepVersion, err := httpd.GetDependencyVersion("2.4.43")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDepVersion := dependency.DepVersion{
					Version:         "2.4.43",
					URI:             "http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:apache:http_server:2.4.43:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/httpd@2.4.43?checksum=some-sha256&download_url=http://archive.apache.org/dist",
					Licenses:        []string{"MIT", "MIT-2"},
				}

				assert.Equal(expectedDepVersion, actualDepVersion)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("http://archive.apache.org/dist/httpd/?F=2&C=M&O=D&P=httpd-2.4.43.tar.bz2*", urlArg)

				urlArg, _ = fakeWebClient.GetArgsForCall(1)
				assert.Equal("http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2.sha1", urlArg)

				urlArg, dependencyPathDownloadArg, _ := fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2", urlArg)

				dependencyPathVerifyArg, checksumArg := fakeChecksummer.VerifySHA1ArgsForCall(0)
				assert.Equal("some-sha1", checksumArg)
				assert.Equal(dependencyPathDownloadArg, dependencyPathVerifyArg)
			})

			when("the SHA1 file is in a different format", func() {
				it("returns the correct httpd version using the SHA1 file", func() {
					index := removeLinesContaining(httpdIndex2443, "httpd-2.4.43.tar.bz2.sha256")
					fakeWebClient.GetReturnsOnCall(0, []byte(index), nil)

					fakeWebClient.GetReturnsOnCall(1, []byte("SHA1(httpd-2.2.14.tar.gz)= some-sha1"), nil)

					fakeChecksummer.GetSHA256Returns("some-sha256", nil)

					depVersion, err := httpd.GetDependencyVersion("2.4.43")
					require.NoError(err)

					assert.Equal("some-sha256", depVersion.SHA256)

					_, checksumArg := fakeChecksummer.VerifySHA1ArgsForCall(0)
					assert.Equal("some-sha1", checksumArg)
				})
			})
		})

		when("there are no SHA256 or SHA1 files", func() {
			it("returns the correct httpd version using the MD5 file", func() {
				index := removeLinesContaining(httpdIndex2443, "httpd-2.4.43.tar.bz2.sha256")
				index = removeLinesContaining(index, "httpd-2.4.43.tar.bz2.sha1")
				fakeWebClient.GetReturnsOnCall(0, []byte(index), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte("some-md5 httpd-2.4.43.tar.bz2"), nil)

				fakeChecksummer.GetSHA256Returns("some-sha256", nil)

				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/httpd@2.4.43?checksum=some-sha256&download_url=http://archive.apache.org/dist")

				actualDepVersion, err := httpd.GetDependencyVersion("2.4.43")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDepVersion := dependency.DepVersion{
					Version:         "2.4.43",
					URI:             "http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:apache:http_server:2.4.43:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/httpd@2.4.43?checksum=some-sha256&download_url=http://archive.apache.org/dist",
					Licenses:        []string{"MIT", "MIT-2"},
				}

				assert.Equal(expectedDepVersion, actualDepVersion)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("http://archive.apache.org/dist/httpd/?F=2&C=M&O=D&P=httpd-2.4.43.tar.bz2*", urlArg)

				urlArg, _ = fakeWebClient.GetArgsForCall(1)
				assert.Equal("http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2.md5", urlArg)

				urlArg, dependencyPathDownloadArg, _ := fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("http://archive.apache.org/dist/httpd/httpd-2.4.43.tar.bz2", urlArg)

				dependencyPathVerifyArg, checksumArg := fakeChecksummer.VerifyMD5ArgsForCall(0)
				assert.Equal("some-md5", checksumArg)
				assert.Equal(dependencyPathDownloadArg, dependencyPathVerifyArg)
			})

			when("the MD5 file in a different format", func() {
				it("returns the correct httpd version using the MD5 file", func() {
					index := removeLinesContaining(httpdIndex2443, "httpd-2.4.43.tar.bz2.sha256")
					index = removeLinesContaining(index, "httpd-2.4.43.tar.bz2.sha1")
					fakeWebClient.GetReturnsOnCall(0, []byte(index), nil)

					fakeWebClient.GetReturnsOnCall(1, []byte("MD5 (httpd-2.4.43.tar.gz) = some-md5"), nil)

					fakeChecksummer.GetSHA256Returns("some-sha256", nil)

					depVersion, err := httpd.GetDependencyVersion("2.4.43")
					require.NoError(err)

					assert.Equal("some-sha256", depVersion.SHA256)

					_, checksumArg := fakeChecksummer.VerifyMD5ArgsForCall(0)
					assert.Equal("some-md5", checksumArg)
				})
			})
		})

		when("the version is known to be missing a checksum", func() {
			it("returns the correct httpd version without verifying the checksum", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(httpdIndex2_2_3), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/httpd@2.2.3?checksum=some-sha256&download_url=http://archive.apache.org/dist")

				actualDepVersion, err := httpd.GetDependencyVersion("2.2.3")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())

				expectedReleaseDate223 := time.Date(2006, 07, 27, 17, 39, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "2.2.3",
					URI:             "http://archive.apache.org/dist/httpd/httpd-2.2.3.tar.bz2",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate223,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:apache:http_server:2.2.3:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/httpd@2.2.3?checksum=some-sha256&download_url=http://archive.apache.org/dist",
					Licenses:        []string{"MIT", "MIT-2"},
				}

				assert.Equal(expectedDepVersion, actualDepVersion)

				assert.Equal(0, fakeChecksummer.VerifyASCCallCount())
				assert.Equal(0, fakeChecksummer.VerifyMD5CallCount())
				assert.Equal(0, fakeChecksummer.VerifySHA1CallCount())
				assert.Equal(0, fakeChecksummer.VerifySHA256CallCount())

				assert.Equal(1, fakeWebClient.GetCallCount())

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("http://archive.apache.org/dist/httpd/?F=2&C=M&O=D&P=httpd-2.2.3.tar.bz2*", urlArg)

				urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("http://archive.apache.org/dist/httpd/httpd-2.2.3.tar.bz2", urlArg)
			})
		})

		when("there are no checksum files", func() {
			it("returns an error", func() {
				index := removeLinesContaining(httpdIndex2443, "httpd-2.4.43.tar.bz2.sha256")
				index = removeLinesContaining(index, "httpd-2.4.43.tar.bz2.sha1")
				index = removeLinesContaining(index, "httpd-2.4.43.tar.bz2.md5")
				index = removeLinesContaining(index, "httpd-2.4.43.tar.bz2.asc")
				fakeWebClient.GetReturnsOnCall(0, []byte(index), nil)

				_, err := httpd.GetDependencyVersion("2.4.43")
				assert.Error(err)
				assert.Contains(err.Error(), "could not find checksum file")
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct httpd release date", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(httpdIndex2443), nil)

			releaseDate, err := httpd.GetReleaseDate("2.4.43")
			require.NoError(err)

			assert.Equal("2020-03-30T14:21:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}

const fullHTTPDIndex = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
 <head>
  <title>Index of /dist/httpd</title>
 </head>
 <body>
<h1>Index of /dist/httpd</h1>

<h2>Apache HTTP Server <u>Source Code</u> Distributions</h2>

<p>This download page includes <strong>only the sources</strong> to compile 
   and build Apache yourself with the proper tools.  Download 
   the precompiled distribution for your platform from 
   <a href="binaries/">binaries/</a>.</p>

<h2>Important Notices</h2>

<ul>
<li><a href="#mirrors">Download from your nearest mirror site!</a></li>
<li><a href="#binaries">Binary Releases</a></li>
<li><a href="#releases">Current Releases</a></li>
<li><a href="#archive">Older Releases</a></li>
<li><a href="#sig">PGP Signatures</a></li>
<li><a href="#patches">Official Patches</a></li>
</ul>

  <table>
   <tr><th valign="top"><img src="/icons/blank.gif" alt="[ICO]"></th><th><a href="?C=N;O=A;F=2;P=httpd-*">Name</a></th><th><a href="?C=M;O=A;F=2;P=httpd-*">Last modified</a></th><th><a href="?C=S;O=A;F=2;P=httpd-*">Size</a></th><th><a href="?C=D;O=A;F=2;P=httpd-*">Description</a></th></tr>
   <tr><th colspan="5"><hr></th></tr>
<tr><td valign="top"><img src="/icons/back.gif" alt="[PARENTDIR]"></td><td><a href="/dist/">Parent Directory</a></td><td>&nbsp;</td><td align="right">  - </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.sha512">httpd-2.4.43.tar.bz2.sha512</a></td><td align="right">2020-03-30 14:21  </td><td align="right">151 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.sha256">httpd-2.4.43.tar.bz2.sha256</a></td><td align="right">2020-03-30 14:21  </td><td align="right"> 87 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.sha1">httpd-2.4.43.tar.bz2.sha1</a></td><td align="right">2020-03-30 14:21  </td><td align="right"> 63 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.md5">httpd-2.4.43.tar.bz2.md5</a></td><td align="right">2020-03-30 14:21  </td><td align="right"> 55 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.asc">httpd-2.4.43.tar.bz2.asc</a></td><td align="right">2020-03-30 14:21  </td><td align="right">488 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.4.43.tar.bz2">httpd-2.4.43.tar.bz2</a></td><td align="right">2020-03-30 14:21  </td><td align="right">6.8M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.4.41.tar.bz2">httpd-2.4.41.tar.bz2</a></td><td align="right">2019-08-12 23:37  </td><td align="right">6.7M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.4.39.tar.bz2">httpd-2.4.39.tar.bz2</a></td><td align="right">2019-03-31 03:42  </td><td align="right">6.7M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.4.38.tar.bz2">httpd-2.4.38.tar.bz2</a></td><td align="right">2019-01-21 15:03  </td><td align="right">6.7M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.2.14.tar.bz2">httpd-2.2.14.tar.bz2</a></td><td align="right">2009-10-03 20:45  </td><td align="right">4.9M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.0.63.tar.bz2">httpd-2.0.63.tar.bz2</a></td><td align="right">2009-10-03 20:45  </td><td align="right">4.4M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.0.61.tar.bz2">httpd-2.0.61.tar.bz2</a></td><td align="right">2007-09-06 19:31  </td><td align="right">4.4M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.2.6.tar.bz2">httpd-2.2.6.tar.bz2</a></td><td align="right">2007-09-06 19:31  </td><td align="right">4.5M</td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.1.6-alpha.tar.bz2">httpd-2.1.6-alpha.tar.bz2</a></td><td align="right">2005-06-24 07:58  </td><td align="right">5.4M</td><td>&nbsp;</td></tr>
   <tr><th colspan="5"><hr></th></tr>
</table>
<h2><a name="mirrors">Download from your
    <a href="http://www.apache.org/dyn/closer.cgi/httpd/"
      >nearest mirror site!</a></a></h2>

<p>Do not download from www.apache.org.  Please use a mirror site
   to help us save apache.org bandwidth.
   <a href="http://www.apache.org/dyn/closer.cgi/httpd/">Go 
   here to find your nearest mirror.</a></p>

<h2><a name="binaries">Binary Releases</a></h2>

<p>Are available in the <a href="binaries/">binaries/</a> directory.
   Every binary distribution contains an install script. See README 
   for details.</p>

<h2><a name="releases">Current Releases</a></h2>

<p>For details on current releases, please see the 
   <a href="http://httpd.apache.org/download.cgi">Apache HTTP
   Server Download Page</a>.</p>

<p>Note; the -win32-src.zip versions of Apache are nearly identical to the
   .tar.gz versions.  However, they offer the source files in DOS/Windows 
   CR/LF text format, and include the Win32 build files.  
   These -win32-src.zip files <strong>do NOT contain binaries!</strong>
   See the <a href="binaries/win32/">binaries/win32/</a> 
   directory for the Windows binary distributions.</p>

<h2><a name="archive">Older Releases</a></h2>

<p>Only current, recommended releases are available on www.apache.org
   and the mirror sites.  Older releases can be obtained from the <a
   href="http://archive.apache.org/dist/httpd/">archive site</a>.</p>

<h2><a name="sig">PGP Signatures</a></h2>

<p>All of the release distribution packages have been digitally signed
   (using PGP or GPG) by the Apache Group members that constructed them.
   There will be an accompanying <SAMP><EM>distribution</EM>.asc</SAMP> file
   in the same directory as the distribution.  The PGP keys can be found
   at the MIT key repository and within this project's
   <a href="http://www.apache.org/dist/httpd/KEYS">KEYS file</a>.</p>

<p>Always use the signature files to verify the authenticity
   of the distribution, <i>e.g.</i>,</p>

<pre>
% pgpk -a KEYS
% pgpv httpd-2.2.8.tar.gz.asc
<i>or</i>,
% pgp -ka KEYS
% pgp httpd-2.2.8.tar.gz.asc
<i>or</i>,
% gpg --import KEYS
% gpg --verify httpd-2.2.8.tar.gz.asc
</pre>

<p>We offer MD5 hashes as an alternative to validate the integrity
   of the downloaded files. A unix program called <code>md5</code> or
   <code>md5sum</code> is included in many unix distributions.  It is
   also available as part of <a
   href="http://www.gnu.org/software/textutils/textutils.html">GNU
   Textutils</a>.  Windows users can get binary md5 programs from <a
   href="http://www.fourmilab.ch/md5/">here</a>, <a
   href="http://www.pc-tools.net/win32/freeware/console/">here</a>, or
   <a href="http://www.slavasoft.com/fsum/">here</a>.</p>

<h2><a name="patches">Official Patches</a></h2>

<p>When we have patches to a minor bug or two, or features which we
   haven't yet included in a new release, we will put them in the
   <A HREF="patches/">patches</A>
   subdirectory so people can get access to it before we roll another
   complete release.</p>
</body></html>
`

const httpdIndex2443 = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
 <head>
  <title>Index of /dist/httpd</title>
 </head>
 <body>
<h1>Index of /dist/httpd</h1>

<h2>Apache HTTP Server <u>Source Code</u> Distributions</h2>

<p>This download page includes <strong>only the sources</strong> to compile 
   and build Apache yourself with the proper tools.  Download 
   the precompiled distribution for your platform from 
   <a href="binaries/">binaries/</a>.</p>

<h2>Important Notices</h2>

<ul>
<li><a href="#mirrors">Download from your nearest mirror site!</a></li>
<li><a href="#binaries">Binary Releases</a></li>
<li><a href="#releases">Current Releases</a></li>
<li><a href="#archive">Older Releases</a></li>
<li><a href="#sig">PGP Signatures</a></li>
<li><a href="#patches">Official Patches</a></li>
</ul>

  <table>
   <tr><th valign="top"><img src="/icons/blank.gif" alt="[ICO]"></th><th><a href="?C=N;O=A;F=2;P=httpd-*">Name</a></th><th><a href="?C=M;O=A;F=2;P=httpd-*">Last modified</a></th><th><a href="?C=S;O=A;F=2;P=httpd-*">Size</a></th><th><a href="?C=D;O=A;F=2;P=httpd-*">Description</a></th></tr>
   <tr><th colspan="5"><hr></th></tr>
<tr><td valign="top"><img src="/icons/back.gif" alt="[PARENTDIR]"></td><td><a href="/dist/">Parent Directory</a></td><td>&nbsp;</td><td align="right">  - </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.sha512">httpd-2.4.43.tar.bz2.sha512</a></td><td align="right">2020-03-30 14:21  </td><td align="right">151 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.sha256">httpd-2.4.43.tar.bz2.sha256</a></td><td align="right">2020-03-30 14:21  </td><td align="right"> 87 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.sha1">httpd-2.4.43.tar.bz2.sha1</a></td><td align="right">2020-03-30 14:21  </td><td align="right"> 63 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.md5">httpd-2.4.43.tar.bz2.md5</a></td><td align="right">2020-03-30 14:21  </td><td align="right"> 55 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.4.43.tar.bz2.asc">httpd-2.4.43.tar.bz2.asc</a></td><td align="right">2020-03-30 14:21  </td><td align="right">488 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.4.43.tar.bz2">httpd-2.4.43.tar.bz2</a></td><td align="right">2020-03-30 14:21  </td><td align="right">6.8M</td><td>&nbsp;</td></tr>
   <tr><th colspan="5"><hr></th></tr>
</table>
<h2><a name="mirrors">Download from your
    <a href="http://www.apache.org/dyn/closer.cgi/httpd/"
      >nearest mirror site!</a></a></h2>

<p>Do not download from www.apache.org.  Please use a mirror site
   to help us save apache.org bandwidth.
   <a href="http://www.apache.org/dyn/closer.cgi/httpd/">Go 
   here to find your nearest mirror.</a></p>

<h2><a name="binaries">Binary Releases</a></h2>

<p>Are available in the <a href="binaries/">binaries/</a> directory.
   Every binary distribution contains an install script. See README 
   for details.</p>

<h2><a name="releases">Current Releases</a></h2>

<p>For details on current releases, please see the 
   <a href="http://httpd.apache.org/download.cgi">Apache HTTP
   Server Download Page</a>.</p>

<p>Note; the -win32-src.zip versions of Apache are nearly identical to the
   .tar.gz versions.  However, they offer the source files in DOS/Windows 
   CR/LF text format, and include the Win32 build files.  
   These -win32-src.zip files <strong>do NOT contain binaries!</strong>
   See the <a href="binaries/win32/">binaries/win32/</a> 
   directory for the Windows binary distributions.</p>

<h2><a name="archive">Older Releases</a></h2>

<p>Only current, recommended releases are available on www.apache.org
   and the mirror sites.  Older releases can be obtained from the <a
   href="http://archive.apache.org/dist/httpd/">archive site</a>.</p>

<h2><a name="sig">PGP Signatures</a></h2>

<p>All of the release distribution packages have been digitally signed
   (using PGP or GPG) by the Apache Group members that constructed them.
   There will be an accompanying <SAMP><EM>distribution</EM>.asc</SAMP> file
   in the same directory as the distribution.  The PGP keys can be found
   at the MIT key repository and within this project's
   <a href="http://www.apache.org/dist/httpd/KEYS">KEYS file</a>.</p>

<p>Always use the signature files to verify the authenticity
   of the distribution, <i>e.g.</i>,</p>

<pre>
% pgpk -a KEYS
% pgpv httpd-2.2.8.tar.gz.asc
<i>or</i>,
% pgp -ka KEYS
% pgp httpd-2.2.8.tar.gz.asc
<i>or</i>,
% gpg --import KEYS
% gpg --verify httpd-2.2.8.tar.gz.asc
</pre>

<p>We offer MD5 hashes as an alternative to validate the integrity
   of the downloaded files. A unix program called <code>md5</code> or
   <code>md5sum</code> is included in many unix distributions.  It is
   also available as part of <a
   href="http://www.gnu.org/software/textutils/textutils.html">GNU
   Textutils</a>.  Windows users can get binary md5 programs from <a
   href="http://www.fourmilab.ch/md5/">here</a>, <a
   href="http://www.pc-tools.net/win32/freeware/console/">here</a>, or
   <a href="http://www.slavasoft.com/fsum/">here</a>.</p>

<h2><a name="patches">Official Patches</a></h2>

<p>When we have patches to a minor bug or two, or features which we
   haven't yet included in a new release, we will put them in the
   <A HREF="patches/">patches</A>
   subdirectory so people can get access to it before we roll another
   complete release.</p>
</body></html>
`

const httpdIndex2_2_3 = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
 <head>
  <title>Index of /dist/httpd</title>
 </head>
 <body>
<h1>Index of /dist/httpd</h1>

<h2>Apache HTTP Server <u>Source Code</u> Distributions</h2>

<p>This download page includes <strong>only the sources</strong> to compile 
   and build Apache yourself with the proper tools.  Download 
   the precompiled distribution for your platform from 
   <a href="binaries/">binaries/</a>.</p>

<h2>Important Notices</h2>

<ul>
<li><a href="#mirrors">Download from your nearest mirror site!</a></li>
<li><a href="#binaries">Binary Releases</a></li>
<li><a href="#releases">Current Releases</a></li>
<li><a href="#archive">Older Releases</a></li>
<li><a href="#sig">PGP Signatures</a></li>
<li><a href="#patches">Official Patches</a></li>
</ul>

  <table>
   <tr><th valign="top"><img src="/icons/blank.gif" alt="[ICO]"></th><th><a href="?C=N;O=A;F=2;P=httpd-*">Name</a></th><th><a href="?C=M;O=A;F=2;P=httpd-*">Last modified</a></th><th><a href="?C=S;O=A;F=2;P=httpd-*">Size</a></th><th><a href="?C=D;O=A;F=2;P=httpd-*">Description</a></th></tr>
   <tr><th colspan="5"><hr></th></tr>
<tr><td valign="top"><img src="/icons/back.gif" alt="[PARENTDIR]"></td><td><a href="/dist/">Parent Directory</a></td><td>&nbsp;</td><td align="right">  - </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/text.gif" alt="[TXT]"></td><td><a href="httpd-2.2.3.tar.bz2.asc">httpd-2.2.3.tar.bz2.asc</a></td><td align="right">2006-07-27 17:39  </td><td align="right">481 </td><td>&nbsp;</td></tr>
<tr><td valign="top"><img src="/icons/unknown.gif" alt="[   ]"></td><td><a href="httpd-2.2.3.tar.bz2">httpd-2.2.3.tar.bz2</a></td><td align="right">2006-07-27 17:39  </td><td align="right">4.7M</td><td>&nbsp;</td></tr>
   <tr><th colspan="5"><hr></th></tr>
</table>
<h2><a name="mirrors">Download from your
    <a href="http://www.apache.org/dyn/closer.cgi/httpd/"
      >nearest mirror site!</a></a></h2>

<p>Do not download from www.apache.org.  Please use a mirror site
   to help us save apache.org bandwidth.
   <a href="http://www.apache.org/dyn/closer.cgi/httpd/">Go 
   here to find your nearest mirror.</a></p>

<h2><a name="binaries">Binary Releases</a></h2>

<p>Are available in the <a href="binaries/">binaries/</a> directory.
   Every binary distribution contains an install script. See README 
   for details.</p>

<h2><a name="releases">Current Releases</a></h2>

<p>For details on current releases, please see the 
   <a href="http://httpd.apache.org/download.cgi">Apache HTTP
   Server Download Page</a>.</p>

<p>Note; the -win32-src.zip versions of Apache are nearly identical to the
   .tar.gz versions.  However, they offer the source files in DOS/Windows 
   CR/LF text format, and include the Win32 build files.  
   These -win32-src.zip files <strong>do NOT contain binaries!</strong>
   See the <a href="binaries/win32/">binaries/win32/</a> 
   directory for the Windows binary distributions.</p>

<h2><a name="archive">Older Releases</a></h2>

<p>Only current, recommended releases are available on www.apache.org
   and the mirror sites.  Older releases can be obtained from the <a
   href="http://archive.apache.org/dist/httpd/">archive site</a>.</p>

<h2><a name="sig">PGP Signatures</a></h2>

<p>All of the release distribution packages have been digitally signed
   (using PGP or GPG) by the Apache Group members that constructed them.
   There will be an accompanying <SAMP><EM>distribution</EM>.asc</SAMP> file
   in the same directory as the distribution.  The PGP keys can be found
   at the MIT key repository and within this project's
   <a href="http://www.apache.org/dist/httpd/KEYS">KEYS file</a>.</p>

<p>Always use the signature files to verify the authenticity
   of the distribution, <i>e.g.</i>,</p>

<pre>
% pgpk -a KEYS
% pgpv httpd-2.2.8.tar.gz.asc
<i>or</i>,
% pgp -ka KEYS
% pgp httpd-2.2.8.tar.gz.asc
<i>or</i>,
% gpg --import KEYS
% gpg --verify httpd-2.2.8.tar.gz.asc
</pre>

<p>We offer MD5 hashes as an alternative to validate the integrity
   of the downloaded files. A unix program called <code>md5</code> or
   <code>md5sum</code> is included in many unix distributions.  It is
   also available as part of <a
   href="http://www.gnu.org/software/textutils/textutils.html">GNU
   Textutils</a>.  Windows users can get binary md5 programs from <a
   href="http://www.fourmilab.ch/md5/">here</a>, <a
   href="http://www.pc-tools.net/win32/freeware/console/">here</a>, or
   <a href="http://www.slavasoft.com/fsum/">here</a>.</p>

<h2><a name="patches">Official Patches</a></h2>

<p>When we have patches to a minor bug or two, or features which we
   haven't yet included in a new release, we will put them in the
   <A HREF="patches/">patches</A>
   subdirectory so people can get access to it before we roll another
   complete release.</p>
</body></html>
`
