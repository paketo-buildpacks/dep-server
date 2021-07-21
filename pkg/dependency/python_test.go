package dependency_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
)

func TestPython(t *testing.T) {
	spec.Run(t, "python", testPython, spec.Report(report.Terminal{}))
}

func testPython(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		python               dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		python, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("python")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all python release versions with the newest first", func() {
			fakeWebClient.GetReturns([]byte(fullPythonIndex), nil)

			versions, err := python.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"3.7.8",
				"3.6.11",
				"3.8.3",
				"2.7.18",
				"3.7.7",
				"3.8.2",
				"3.8.1",
				"3.7.6",
				"3.6.10",
				"3.5.9",
				"3.5.8",
				"2.7.17",
				"3.7.5",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://www.python.org/downloads/", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct python version", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(python378DownloadPage), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)

			fakeChecksummer.GetSHA256Returns("some-sha256", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/python@3.7.8?checksum=some-sha-256&download_url=https://www.python.org")

			actualDepVersion, err := python.GetDependencyVersion("3.7.8")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC)
			expectedDeprecationDate := time.Date(2023, 6, 27, 0, 0, 0, 0, time.UTC)
			expectedDepVersion := dependency.DepVersion{
				Version:         "3.7.8",
				URI:             "https://www.python.org/ftp/python/3.7.8/Python-3.7.8.tgz",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: &expectedDeprecationDate,
				CPE:             "cpe:2.3:a:python:python:3.7.8:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/python@3.7.8?checksum=some-sha-256&download_url=https://www.python.org",
				Licenses:        []string{"MIT", "MIT-2"},
			}

			assert.Equal(expectedDepVersion, actualDepVersion)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://www.python.org/downloads/release/python-378/", urlArg)

			_, checksumArg := fakeChecksummer.VerifyMD5ArgsForCall(0)
			assert.Equal("4d5b16e8c15be38eb0f4b8f04eb68cd0", checksumArg)
		})

		when("the release date uses an abbreviated month", func() {
			it("assumes the first day of the month", func() {
				downloadPage := strings.ReplaceAll(python378DownloadPage, "June 27, 2020", "Sept 2, 2020")
				fakeWebClient.GetReturnsOnCall(0, []byte(downloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)

				depVersion, err := python.GetDependencyVersion("3.7.8")
				require.NoError(err)

				assert.Equal("2020-09-02T00:00:00Z", depVersion.ReleaseDate.Format(time.RFC3339))
			})
		})

		when("the deprecation date does not include a day", func() {
			it("assumes the first day of the month", func() {
				index := strings.ReplaceAll(fullPythonIndex, "2023-06-27", "2023-06")
				fakeWebClient.GetReturnsOnCall(0, []byte(python378DownloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(index), nil)

				depVersion, err := python.GetDependencyVersion("3.7.8")
				require.NoError(err)

				assert.Equal("2023-06-01T00:00:00Z", depVersion.DeprecationDate.Format(time.RFC3339))
			})
		})

		when("the deprecation date is missing", func() {
			it("leaves it empty", func() {
				index := strings.ReplaceAll(fullPythonIndex, ">3.7<", "><")
				fakeWebClient.GetReturnsOnCall(0, []byte(python378DownloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(index), nil)

				depVersion, err := python.GetDependencyVersion("3.7.8")
				require.NoError(err)

				assert.Empty(depVersion.DeprecationDate)
			})
		})

		when("tarball is spelled as two words", func() {
			it("still finds the source URI", func() {
				index := strings.ReplaceAll(fullPythonIndex, "Gzipped source tarball", "Gzipped source tar ball")
				fakeWebClient.GetReturnsOnCall(0, []byte(python378DownloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(index), nil)

				depVersion, err := python.GetDependencyVersion("3.7.8")
				require.NoError(err)

				assert.Equal("https://www.python.org/ftp/python/3.7.8/Python-3.7.8.tgz", depVersion.URI)
			})
		})

		when("the md5 in the file list is wrong but the one in the pre block is correct", func() {
			it("uses the one from the pre block", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(python333DownloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeChecksummer.VerifyMD5ReturnsOnCall(0, errors.New("some-error"))

				depVersion, err := python.GetDependencyVersion("3.3.3")
				require.NoError(err)

				assert.Equal("some-sha256", depVersion.SHA256)

				_, firstMD5 := fakeChecksummer.VerifyMD5ArgsForCall(0)
				_, secondMD5 := fakeChecksummer.VerifyMD5ArgsForCall(1)

				assert.ElementsMatch(
					[]string{firstMD5, secondMD5},
					[]string{"831d59212568dc12c95df222865d3441", "a44bec5d1391b1af654cf15e25c282f2"},
				)
			})
		})

		when("the md5 in the file list is wrong but the one in the blockquote is correct", func() {
			it("uses the one from the pre block", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(python255DownloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeChecksummer.VerifyMD5ReturnsOnCall(0, errors.New("some-error"))

				depVersion, err := python.GetDependencyVersion("2.5.5")
				require.NoError(err)

				assert.Equal("some-sha256", depVersion.SHA256)

				_, firstMD5 := fakeChecksummer.VerifyMD5ArgsForCall(0)
				_, secondMD5 := fakeChecksummer.VerifyMD5ArgsForCall(1)

				assert.ElementsMatch(
					[]string{firstMD5, secondMD5},
					[]string{"abc02139ca38f4258e8e372f7da05c88", "6953d49c4d2470d88d8577b4e5ed3ce2"},
				)
			})
		})

		when("the MD5s are known to be wrong", func() {
			it("does not try to verify the MD5", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(python255DownloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)

				depVersion, err := python.GetDependencyVersion("3.1.0")
				require.NoError(err)

				assert.Equal("some-sha256", depVersion.SHA256)

				assert.Equal(0, fakeChecksummer.VerifyMD5CallCount())
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct python release date", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(python378DownloadPage), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)
			fakeChecksummer.GetSHA256Returns("some-sha256", nil)

			releaseDate, err := python.GetReleaseDate("3.7.8")
			require.NoError(err)

			assert.Equal("2020-06-27T00:00:00Z", releaseDate.Format(time.RFC3339))
		})

		when("the release date uses an abbreviated month", func() {
			it("assumes the first day of the month", func() {
				downloadPage := strings.ReplaceAll(python378DownloadPage, "June 27, 2020", "Sept 2, 2020")
				fakeWebClient.GetReturnsOnCall(0, []byte(downloadPage), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(fullPythonIndex), nil)

				releaseDate, err := python.GetReleaseDate("3.7.8")
				require.NoError(err)

				assert.Equal("2020-09-02T00:00:00Z", releaseDate.Format(time.RFC3339))
			})
		})
	})
}

const fullPythonIndex = `<!doctype html>
<html class="no-js" lang="en" dir="ltr">
<body class="python download">
    <div id="touchnav-wrapper">
        <div id="nojs" class="do-not-print">
            <p><strong>Notice:</strong> While Javascript is not essential for this website, your interaction with the content will be limited. Please turn Javascript on for the full experience. </p>
        </div>
        <div id="content" class="content-wrapper">
            <!-- Main Content Column -->
            <div class="container">
                <section class="main-content " role="main">
                <div class="row active-release-list-widget">
                    <h2 class="widget-title">Active Python Releases</h2>
                    <p class="success-quote"><a href="https://devguide.python.org/#status-of-python-branches">For more information visit the Python Developer's Guide</a>.</p>
                    <div class="list-row-headings">
                        <span class="release-version">Python version</span>
                        <span class="release-status">Maintenance status</span>
                        <span class="release-start">First released</span>
                        <span class="release-end">End of support</span>
                        <span class="release-pep">Release schedule</span>
                    </div>
                    <ol class="list-row-container menu">
                        <li>
                            <span class="release-version">3.8</span>
                            <span class="release-status">bugfix</span>
                            <span class="release-start">2019-10-14</span>
                            <span class="release-end">2024-10</span>
                            <span class="release-pep"><a href="https://www.python.org/dev/peps/pep-0569">PEP 569</a></span>
                        </li>
                        <li>
                            <span class="release-version">3.7</span>
                            <span class="release-status">bugfix</span>
                            <span class="release-start">2018-06-27</span>
                            <span class="release-end">2023-06-27</span>
                            <span class="release-pep"><a href="https://www.python.org/dev/peps/pep-0537">PEP 537</a></span>
                        </li>
                        <li>
                            <span class="release-version">3.6</span>
                            <span class="release-status">security</span>
                            <span class="release-start">2016-12-23</span>
                            <span class="release-end">2021-12-23</span>
                            <span class="release-pep"><a href="https://www.python.org/dev/peps/pep-0494">PEP 494</a></span>
                        </li>
                        <li>
                            <span class="release-version">3.5</span>
                            <span class="release-status">security</span>
                            <span class="release-start">2015-09-13</span>
                            <span class="release-end">2020-09-13</span>
                            <span class="release-pep"><a href="https://www.python.org/dev/peps/pep-0478">PEP 478</a></span>
                        </li>
                        <li>
                            <span class="release-version">2.7</span>
                            <span class="release-status">end-of-life</span>
                            <span class="release-start">2010-07-03</span>
                            <span class="release-end">2020-01-01</span>
                            <span class="release-pep"><a href="https://www.python.org/dev/peps/pep-0373">PEP 373</a></span>
                        </li>
                    </ol>
                </div>
                <div class="row download-list-widget">
                    <h2 class="widget-title">Looking for a specific release?</h2>
                    <p class="success-quote">Python releases by version number:</p>
                    <div class="list-row-headings">
                        <span class="release-number">Release version</span>
                        <span class="release-date">Release date</span>
                        <span class="release-download">&nbsp;</span>
                        <span class="release-enhancements">Click for more</span>
                    </div>
                    <ol class="list-row-container menu">
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-378/">Python 3.7.8</a></span>
                            <span class="release-date">June 27, 2020</span>
                            <span class="release-download"><a href="/downloads/release/python-378/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/release/3.7.8/whatsnew/changelog.html#changelog">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-3611/">Python 3.6.11</a></span>
                            <span class="release-date">June 27, 2020</span>
                            <span class="release-download"><a href="/downloads/release/python-3611/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/release/3.6.11/whatsnew/changelog.html#changelog">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-383/">Python 3.8.3</a></span>
                            <span class="release-date">May 13, 2020</span>
                            <span class="release-download"><a href="/downloads/release/python-383/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/release/3.8.3/whatsnew/changelog.html#changelog">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-2718/">Python 2.7.18</a></span>
                            <span class="release-date">April 20, 2020</span>
                            <span class="release-download"><a href="/downloads/release/python-2718/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-377/">Python 3.7.7</a></span>
                            <span class="release-date">March 10, 2020</span>
                            <span class="release-download"><a href="/downloads/release/python-377/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.7/whatsnew/changelog.html#python-3-7-7-final">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-382/">Python 3.8.2</a></span>
                            <span class="release-date">Feb. 24, 2020</span>
                            <span class="release-download"><a href="/downloads/release/python-382/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/release/3.8.2/whatsnew/changelog.html#python-3-8-2-final">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-381/">Python 3.8.1</a></span>
                            <span class="release-date">Dec. 18, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-381/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.8/whatsnew/changelog.html#python-3-8-1">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-376/">Python 3.7.6</a></span>
                            <span class="release-date">Dec. 18, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-376/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.7/whatsnew/changelog.html#python-3-7-6-final">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-3610/">Python 3.6.10</a></span>
                            <span class="release-date">Dec. 18, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-3610/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.6/whatsnew/changelog.html#python-3-6-10-final">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-359/">Python 3.5.9</a></span>
                            <span class="release-date">Nov. 2, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-359/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.5/whatsnew/changelog.html#python-3-5-9">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-358/">Python 3.5.8</a></span>
                            <span class="release-date">Oct. 29, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-358/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.5/whatsnew/changelog.html#python-3-5-8">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-2717/">Python 2.7.17</a></span>
                            <span class="release-date">Oct. 19, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-2717/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://raw.githubusercontent.com/python/cpython/c2f86d86e6c8f5fd1ef602128b537a48f3f5c063/Misc/NEWS.d/2.7.17rc1.rst">Release Notes</a></span>
                        </li>
                        <li>
                            <span class="release-number"><a href="/downloads/release/python-375/">Python 3.7.5</a></span>
                            <span class="release-date">Oct. 15, 2019</span>
                            <span class="release-download"><a href="/downloads/release/python-375/"><span aria-hidden="true" class="icon-download"></span> Download</a></span>
                            <span class="release-enhancements"><a href="https://docs.python.org/3.7/whatsnew/changelog.html#python-3-7-5-final">Release Notes</a></span>
                        </li>
                    </ol>
                    <p><a href="/download/releases/">View older releases</a><!-- removed by Frank until content available <small><em>Older releases: <a href="#">Source releases, <a href="#">binaries-1.1</a>, <a href="#">binaries-1.2</a>, <a href="#">binaries-1.3</a>, <a href="#">binaries-1.4</a>, <a href="#">binaries-1.5</a></em></small> --></p>
                </div>
                <div class="row">
                    <div class="small-widget download-widget1">
                        <h2 class="widget-title">Licenses</h2>
<p>All Python releases are <a href="http://www.opensource.org/">Open Source</a>. Historically, most, but not all, Python releases have also been GPL-compatible. The Licenses page details GPL-compatibility and Terms and Conditions. </p>
<p><a class="readmore" href="http://docs.python.org/3/license.html#terms-and-conditions-for-accessing-or-otherwise-using-python">Read more</a></p>
                    </div>
                    <div class="small-widget download-widget2">
                        <h2 class="widget-title">Sources</h2>
<p>For most Unix systems, you must download and compile the source code. The same source code archive can also be used to build the Windows and Mac versions, and is the starting point for ports to all other platforms.</p>
<p>Download the latest <a href="https://www.python.org/ftp/python/3.8.3/Python-3.8.3.tar.xz">Python 3</a> and <a href="https://www.python.org/ftp/python/2.7.18/Python-2.7.18.tar.xz">Python 2</a> source.</p>
<p><a class="readmore" href="/download/source/">Read more</a></p>
                    </div>
                    <div class="small-widget download-widget3">
                        <h2 class="widget-title">Alternative Implementations</h2>
<p>This site hosts the "traditional" implementation of Python (nicknamed CPython). A number of alternative implementations are available as well. </p>
<p><a class="readmore" href="/download/alternatives/">Read more</a></p>
                    </div>
                    <div class="small-widget download-widget3 last">
                        <h2 class="widget-title">History</h2>
<p>Python was created in the early 1990s by Guido van Rossum at Stichting Mathematisch Centrum in the Netherlands as a successor of a language called ABC. Guido remains Python’s principal author, although it includes many contributions from others. </p>
<p><a class="readmore" href="http://docs.python.org/3/license.html">Read more</a></p>
                    </div>
                </div>
                </div>
                </section>
            </div><!-- end .container -->
        </div><!-- end #content .content-wrapper -->
    </div><!-- end #touchnav-wrapper -->
</body>
</html>`

const python378DownloadPage = `<!doctype html>
<html class="no-js" lang="en" dir="ltr">
<body class="python downloads">
    <div id="touchnav-wrapper">
        <div id="content" class="content-wrapper">
            <!-- Main Content Column -->
            <div class="container">
                <section class="main-content " role="main">
    <article class="text">
        <header class="article-header">
            <h1 class="page-title">Python 3.7.8</h1>
        </header>
        <p><strong>Release Date:</strong> June 27, 2020</p>
        <table>
          <thead>
            <tr>
              <th>Version</th>
              <th>Operating System</th>
              <th>Description</th>
              <th>MD5 Sum</th>
              <th>File Size</th>
              <th>GPG</th>
            </tr>
          </thead>
          <tbody>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/Python-3.7.8.tgz">Gzipped source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>4d5b16e8c15be38eb0f4b8f04eb68cd0</td>
                <td>23276116</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/Python-3.7.8.tgz.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/Python-3.7.8.tar.xz">XZ compressed source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>a224ef2249a18824f48fba9812f4006f</td>
                <td>17399552</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/Python-3.7.8.tar.xz.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-macosx10.9.pkg">macOS 64-bit installer</a></td>
                <td>Mac OS X</td>
                <td>for OS X 10.9 and later</td>
                <td>2819435f3144fd973d3dea4ae6969f6d</td>
                <td>29303677</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-macosx10.9.pkg.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python378.chm">Windows help file</a></td>
                <td>Windows</td>
                <td></td>
                <td>65bb54986e5a921413e179d2211b9bfb</td>
                <td>8186659</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python378.chm.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-embed-amd64.zip">Windows x86-64 embeddable zip file</a></td>
                <td>Windows</td>
                <td>for AMD64/EM64T/x64</td>
                <td>5ae191973e00ec490cf2a93126ce4d89</td>
                <td>7536190</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-embed-amd64.zip.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-amd64.exe">Windows x86-64 executable installer</a></td>
                <td>Windows</td>
                <td>for AMD64/EM64T/x64</td>
                <td>70b08ab8e75941da7f5bf2b9be58b945</td>
                <td>26993432</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-amd64.exe.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-amd64-webinstall.exe">Windows x86-64 web-based installer</a></td>
                <td>Windows</td>
                <td>for AMD64/EM64T/x64</td>
                <td>b07dbb998a4a0372f6923185ebb6bf3e</td>
                <td>1363056</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-amd64-webinstall.exe.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-embed-win32.zip">Windows x86 embeddable zip file</a></td>
                <td>Windows</td>
                <td></td>
                <td>5f0f83433bd57fa55182cb8ea42d43d6</td>
                <td>6765162</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-embed-win32.zip.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8.exe">Windows x86 executable installer</a></td>
                <td>Windows</td>
                <td></td>
                <td>4a9244c57f61e3ad2803e900a2f75d77</td>
                <td>25974352</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8.exe.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-webinstall.exe">Windows x86 web-based installer</a></td>
                <td>Windows</td>
                <td></td>
                <td>642e566f4817f118abc38578f3cc4e69</td>
                <td>1324944</td>
                <td><a href="https://www.python.org/ftp/python/3.7.8/python-3.7.8-webinstall.exe.asc">SIG</a></td>
              </tr>
          </tbody>
        </table>
    </article>
                </section>
            </div><!-- end .container -->
        </div><!-- end #content .content-wrapper -->
    </div><!-- end #touchnav-wrapper -->
</body>
</html>`

const python333DownloadPage = `<!doctype html>
<html class="no-js" lang="en" dir="ltr">  <!--<![endif]-->
<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <link rel="prefetch" href="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js">
    <meta name="application-name" content="Python.org">
    <meta name="msapplication-tooltip" content="The official home of the Python Programming Language">
    <meta name="apple-mobile-web-app-title" content="Python.org">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="HandheldFriendly" content="True">
    <meta name="format-detection" content="telephone=no">
    <meta http-equiv="cleartype" content="on">
    <meta http-equiv="imagetoolbar" content="false">
    <script src="/static/js/libs/modernizr.js"></script>
    <link href="/static/stylesheets/style.30afed881237.css" rel="stylesheet" type="text/css" title="default" />
    <link href="/static/stylesheets/mq.eef77a5d2257.css" rel="stylesheet" type="text/css" media="not print, braille, embossed, speech, tty" />
    <!--[if (lte IE 8)&(!IEMobile)]>
    <link href="/static/stylesheets/no-mq.7946159eb289.css" rel="stylesheet" type="text/css" media="screen" />
    <![endif]-->
    <link rel="icon" type="image/x-icon" href="/static/favicon.ico">
    <link rel="apple-touch-icon-precomposed" sizes="144x144" href="/static/apple-touch-icon-144x144-precomposed.png">
    <link rel="apple-touch-icon-precomposed" sizes="114x114" href="/static/apple-touch-icon-114x114-precomposed.png">
    <link rel="apple-touch-icon-precomposed" sizes="72x72" href="/static/apple-touch-icon-72x72-precomposed.png">
    <link rel="apple-touch-icon-precomposed" href="/static/apple-touch-icon-precomposed.png">
    <link rel="apple-touch-icon" href="/static/apple-touch-icon-precomposed.png">
    <meta name="msapplication-TileImage" content="/static/metro-icon-144x144-precomposed.png"><!-- white shape -->
    <meta name="msapplication-TileColor" content="#3673a5"><!-- python blue -->
    <meta name="msapplication-navbutton-color" content="#3673a5">
    <title>Python Release Python 3.3.3 | Python.org</title>
    <meta name="description" content="The official home of the Python Programming Language">
    <meta name="keywords" content="Python programming language object oriented web free open source software license documentation download community">
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="Python.org">
    <meta property="og:title" content="Python Release Python 3.3.3">
    <meta property="og:description" content="The official home of the Python Programming Language">
    <meta property="og:image" content="https://www.python.org/static/opengraph-icon-200x200.png">
    <meta property="og:image:secure_url" content="https://www.python.org/static/opengraph-icon-200x200.png">
    <meta property="og:url" content="https://www.python.org/downloads/release/python-333/">
    <link rel="author" href="/static/humans.txt">
    <link rel="alternate" type="application/rss+xml" title="Python Enhancement Proposals"
          href="https://www.python.org/dev/peps/peps.rss/">
    <link rel="alternate" type="application/rss+xml" title="Python Job Opportunities"
          href="https://www.python.org/jobs/feed/rss/">
    <link rel="alternate" type="application/rss+xml" title="Python Software Foundation News"
          href="https://feeds.feedburner.com/PythonSoftwareFoundationNews">
    <link rel="alternate" type="application/rss+xml" title="Python Insider"
          href="https://feeds.feedburner.com/PythonInsider">
    <script type="application/ld+json">
     {
       "@context": "https://schema.org",
       "@type": "WebSite",
       "url": "https://www.python.org/",
       "potentialAction": {
         "@type": "SearchAction",
         "target": "https://www.python.org/search/?q={search_term_string}",
         "query-input": "required name=search_term_string"
       }
     }
    </script>
    <script type="text/javascript">
    var _gaq = _gaq || [];
    _gaq.push(['_setAccount', 'UA-39055973-1']);
    _gaq.push(['_trackPageview']);
    (function() {
        var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
        ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
        var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
    })();
    </script>
</head>
<body class="python downloads">
    <div id="touchnav-wrapper">
        <div id="content" class="content-wrapper">
            <!-- Main Content Column -->
            <div class="container">
                <section class="main-content " role="main">
<ul class="breadcrumbs menu">
</ul>
    <article class="text">
        <header class="article-header">
            <h1 class="page-title">Python 3.3.3</h1>
        </header>
        <p><strong>Release Date:</strong> Nov. 17, 2013</p>
        <!-- Migrated from Release.release_page field. -->
<p>fixes <a class="reference external" href="http://docs.python.org/3.3/whatsnew/changelog.html">several security issues and various other bugs</a> found in Python
3.3.2.</p>
<p>This release fully supports OS X 10.9 Mavericks.  In particular, this release
fixes an issue that could cause previous versions of Python to crash when typing
in interactive mode on OS X 10.9.</p>
<div class="section" id="major-new-features-of-the-3-3-series-compared-to-3-2">
<h1>Major new features of the 3.3 series, compared to 3.2</h1>
<p>Python 3.3 includes a range of improvements of the 3.x series, as well as easier
porting between 2.x and 3.x.</p>
<ul class="simple">
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0380">PEP 380</a>, syntax for delegating to a subgenerator (<tt class="docutils literal">yield from</tt>)</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0393">PEP 393</a>, flexible string representation (doing away with the distinction
between &quot;wide&quot; and &quot;narrow&quot; Unicode builds)</li>
<li>A C implementation of the &quot;decimal&quot; module, with up to 120x speedup
for decimal-heavy applications</li>
<li>The import system (__import__) is based on importlib by default</li>
<li>The new &quot;lzma&quot; module with LZMA/XZ support</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0397">PEP 397</a>, a Python launcher for Windows</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0405">PEP 405</a>, virtual environment support in core</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0420">PEP 420</a>, namespace package support</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-3151">PEP 3151</a>, reworking the OS and IO exception hierarchy</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-3155">PEP 3155</a>, qualified name for classes and functions</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0409">PEP 409</a>, suppressing exception context</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0414">PEP 414</a>, explicit Unicode literals to help with porting</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0418">PEP 418</a>, extended platform-independent clocks in the &quot;time&quot; module</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0412">PEP 412</a>, a new key-sharing dictionary implementation that significantly
saves memory for object-oriented code</li>
<li><a class="reference external" href="http://www.python.org/dev/peps/pep-0362">PEP 362</a>, the function-signature object</li>
<li>The new &quot;faulthandler&quot; module that helps diagnosing crashes</li>
<li>The new &quot;unittest.mock&quot; module</li>
<li>The new &quot;ipaddress&quot; module</li>
<li>The &quot;sys.implementation&quot; attribute</li>
<li>A policy framework for the email package, with a provisional (see
<a class="reference external" href="http://www.python.org/dev/peps/pep-0411">PEP 411</a>) policy that adds much improved unicode support for email
header parsing</li>
<li>A &quot;collections.ChainMap&quot; class for linking mappings to a single unit</li>
<li>Wrappers for many more POSIX functions in the &quot;os&quot; and &quot;signal&quot; modules, as
well as other useful functions such as &quot;sendfile()&quot;</li>
<li>Hash randomization, introduced in earlier bugfix releases, is now
switched on by default</li>
</ul>
</div>
<div class="section" id="more-resources">
<h1>More resources</h1>
<ul class="simple">
<li><a class="reference external" href="http://docs.python.org/3.3/whatsnew/changelog.html">Change log for this release</a>.</li>
<li><a class="reference external" href="http://docs.python.org/3.3/">Online Documentation</a></li>
<li><a class="reference external" href="http://docs.python.org/3.3/whatsnew/3.3.html">What's new in 3.3?</a></li>
<li><a class="reference external" href="http://python.org/dev/peps/pep-0398/">3.3 Release Schedule</a></li>
<li>Report bugs at <a class="reference external" href="http://bugs.python.org">http://bugs.python.org</a>.</li>
<li><a class="reference external" href="/psf/donations/">Help fund Python and its community</a>.</li>
</ul>
<div class="section" id="download">
<h2>Download</h2>
<!-- This is a preview release, and its use is not recommended in production
settings. -->
<p>This is a production release.  Please <a class="reference external" href="http://bugs.python.org">report any bugs</a> you encounter.</p>
<p>We currently support these formats for download:</p>
<ul class="simple">
<li><a class="reference external" href="/ftp/python/3.3.3/Python-3.3.3.tar.bz2">Bzipped source tar ball (3.3.3)</a> <a class="reference external" href="/ftp/python/3.3.3/Python-3.3.3.tar.bz2.asc">(sig)</a>, ~ 14 MB</li>
<li><a class="reference external" href="/ftp/python/3.3.3/Python-3.3.3.tar.xz">XZ compressed source tar ball (3.3.3)</a>
<a class="reference external" href="/ftp/python/3.3.3/Python-3.3.3.tar.xz.asc">(sig)</a>, ~ 11 MB</li>
<li><a class="reference external" href="/ftp/python/3.3.3/Python-3.3.3.tgz">Gzipped source tar ball (3.3.3)</a> <a class="reference external" href="/ftp/python/3.3.3/Python-3.3.3.tgz.asc">(sig)</a>, ~ 16 MB</li>
</ul>
<!--  -->
<!-- Windows binaries will be provided shortly. -->
<ul class="simple">
<li><a class="reference external" href="/ftp/python/3.3.3/python-3.3.3.msi">Windows x86 MSI Installer (3.3.3)</a> <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3.msi.asc">(sig)</a> and <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3-pdb.zip">Visual Studio debug information
files</a> <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3-pdb.zip.asc">(sig)</a></li>
<li><a class="reference external" href="/ftp/python/3.3.3/python-3.3.3.amd64.msi">Windows X86-64 MSI Installer (3.3.3)</a>
<a class="footnote-reference" href="#id4" id="id1">[1]</a> <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3.amd64.msi.asc">(sig)</a> and <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3.amd64-pdb.zip">Visual Studio
debug information files</a> <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3.amd64-pdb.zip.asc">(sig)</a></li>
<li><a class="reference external" href="/ftp/python/3.3.3/python333.chm">Windows help file</a> <a class="reference external" href="/ftp/python/3.3.3/python333.chm.asc">(sig)</a></li>
</ul>
<!--  -->
<!-- Mac binaries will be provided shortly. -->
<ul class="simple">
<li><a class="reference external" href="/ftp/python/3.3.3/python-3.3.3-macosx10.6.dmg">Mac OS X 64-bit/32-bit Installer (3.3.3) for Mac OS X 10.6 and later</a> <a class="footnote-reference" href="#id5" id="id2">[2]</a> <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3-macosx10.6.dmg.asc">(sig)</a>.
[You may need an updated Tcl/Tk install to run IDLE or use Tkinter,
see note 2 for instructions.]</li>
<li><a class="reference external" href="/ftp/python/3.3.3/python-3.3.3-macosx10.5.dmg">Mac OS X 32-bit i386/PPC Installer (3.3.3) for Mac OS X 10.5 and later</a> <a class="footnote-reference" href="#id5" id="id3">[2]</a> <a class="reference external" href="/ftp/python/3.3.3/python-3.3.3-macosx10.5.dmg.asc">(sig)</a></li>
</ul>
<p>The source tarballs are signed with Georg Brandl's key, which has a key id of
36580288; the fingerprint is <tt class="docutils literal">26DE A9D4 6133 91EF 3E25 C9FF 0A5B 1018 3658
0288</tt>. The Windows installer was signed by Martin von Löwis' public key, which
has a key id of 7D9DC8D2.  The Mac installers were signed with Ned Deily's key,
which has a key id of 6F5E1540.  The public keys are located on the <a class="reference external" href="/download#pubkeys">download
page</a>.</p>
<p>MD5 checksums and sizes of the released files:</p>
<pre class="literal-block">
831d59212568dc12c95df222865d3441  16808057  Python-3.3.3.tgz
f3ebe34d4d8695bf889279b54673e10c  14122529  Python-3.3.3.tar.bz2
4ca001c5586eb0744e3174bc75c6fba8  12057744  Python-3.3.3.tar.xz
60f44c22bbd00fbf3f63d98ef761295b  19876666  python-3.3.3-macosx10.5.dmg
3f7b6c1dc58d7e0b5282f3b7a2e00ef7  19956580  python-3.3.3-macosx10.6.dmg
3fc2925746372ab8401dfabce278d418  27034152  python-3.3.3-pdb.zip
8af44d33ea3a1528fc56b3a362924500  22145398  python-3.3.3.amd64-pdb.zip
8de52d1e2e4bbb3419b7f40bdf48e855  21086208  python-3.3.3.amd64.msi
ab6a031aeca66507e4c8697ff93a0007  20537344  python-3.3.3.msi
c86d6d68ca1a1de7395601a4918314f9   6651185  python333.chm
</pre>
<table class="docutils footnote" frame="void" id="id4" rules="none">
<colgroup><col class="label" /><col /></colgroup>
<tbody valign="top">
<tr><td class="label"><a class="fn-backref" href="#id1">[1]</a></td><td>The binaries for AMD64 will also work on processors that implement the
Intel 64 architecture (formerly EM64T), i.e. the architecture that
Microsoft calls x64, and AMD called x86-64 before calling it AMD64. They
will not work on Intel Itanium Processors (formerly IA-64).</td></tr>
</tbody>
</table>
<table class="docutils footnote" frame="void" id="id5" rules="none">
<colgroup><col class="label" /><col /></colgroup>
<tbody valign="top">
<tr><td class="label">[2]</td><td><em>(<a class="fn-backref" href="#id2">1</a>, <a class="fn-backref" href="#id3">2</a>)</em> There is <a class="reference external" href="/download/mac/tcltk">important information about IDLE, Tkinter, and Tcl/Tk on Mac OS
X here</a>.</td></tr>
</tbody>
</table>
</div>
</div>
        <p><a href="https://docs.python.org/release/3.3.3/whatsnew/changelog.html">Full Changelog</a></p>
        <header class="article-header">
            <h1 class="page-title">Files</h1>
        </header>
        <table>
          <thead>
            <tr>
              <th>Version</th>
              <th>Operating System</th>
              <th>Description</th>
              <th>MD5 Sum</th>
              <th>File Size</th>
              <th>GPG</th>
            </tr>
          </thead>
          <tbody>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/Python-3.3.3.tar.bz2">bzip2 compressed source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>f3ebe34d4d8695bf889279b54673e10c</td>
                <td>14122529</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/Python-3.3.3.tar.bz2.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/Python-3.3.3.tgz">Gzipped source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>a44bec5d1391b1af654cf15e25c282f2</td>
                <td>69120000</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/Python-3.3.3.tgz.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/Python-3.3.3.tar.xz">XZ compressed source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>4ca001c5586eb0744e3174bc75c6fba8</td>
                <td>12057744</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/Python-3.3.3.tar.xz.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3-macosx10.5.dmg">Mac OS X 32-bit i386/PPC installer</a></td>
                <td>Mac OS X</td>
                <td>for Mac OS X 10.5 and later</td>
                <td>60f44c22bbd00fbf3f63d98ef761295b</td>
                <td>19876666</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3-macosx10.5.dmg.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3-macosx10.6.dmg">Mac OS X 64-bit/32-bit installer</a></td>
                <td>Mac OS X</td>
                <td>for Mac OS X 10.6 and later</td>
                <td>3f7b6c1dc58d7e0b5282f3b7a2e00ef7</td>
                <td>19956580</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3-macosx10.6.dmg.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3-pdb.zip">Windows debug information files</a></td>
                <td>Windows</td>
                <td></td>
                <td>3fc2925746372ab8401dfabce278d418</td>
                <td>27034152</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3-pdb.zip.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python333.chm">Windows help file</a></td>
                <td>Windows</td>
                <td></td>
                <td>c86d6d68ca1a1de7395601a4918314f9</td>
                <td>6651185</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python333.chm.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3.amd64.msi">Windows x86-64 MSI installer</a></td>
                <td>Windows</td>
                <td>for AMD64/EM64T/x64</td>
                <td>8de52d1e2e4bbb3419b7f40bdf48e855</td>
                <td>21086208</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3.amd64.msi.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3.msi">Windows x86 MSI installer</a></td>
                <td>Windows</td>
                <td></td>
                <td>ab6a031aeca66507e4c8697ff93a0007</td>
                <td>20537344</td>
                <td><a href="https://www.python.org/ftp/python/3.3.3/python-3.3.3.msi.asc">SIG</a></td>
              </tr>
          </tbody>
        </table>
    </article>
                </section>
            </div><!-- end .container -->
        </div><!-- end #content .content-wrapper -->
    </div><!-- end #touchnav-wrapper -->
</body>
</html>`

const python255DownloadPage = `<html class="no-js" lang="en" dir="ltr">  <!--<![endif]-->
<body class="python downloads">
    <div id="touchnav-wrapper">
        <div id="content" class="content-wrapper">
            <!-- Main Content Column -->
            <div class="container">
                <section class="main-content " role="main">
<ul class="breadcrumbs menu">
</ul>
    <article class="text">
        <header class="article-header">
            <h1 class="page-title">Python 2.5.5</h1>
        </header>
        <p><strong>Release Date:</strong> Jan. 31, 2010</p>
        <!-- Migrated from Release.release_page field.
We are pleased to announce the release of
**Python 2.5.5**, a
security fix release of Python 2.5, on January 31st, 2010.
**Python 2.5.5 has been replaced by a newer bugfix
release of Python**. Please download 'Python 2.5.6 <../2.5.6/>'__ instead.
The last binary release of Python 2.5 was '2.5.4 <../2.5.4/>'_. -->
<p>This is a source-only release that only includes security
fixes. The last full bug-fix release of Python 2.5 was
<a class="reference external" href="../2.5.4/">Python 2.5.4</a>. User are encouraged to upgrade
to the latest release of Python 2.7 (which is <a class="reference external" href="../2.7.2/">2.7.2</a>
at this point).</p>
<p>This releases fixes issues with the logging, tarfile and expat
modules, and with thread-local variables. See the <a class="reference external" href="NEWS.txt">detailed release
notes</a> for more details.</p>
<p>See also the <a class="reference external" href="license">license</a>.</p>
<div class="section" id="download-the-release">
<h1>Download the release</h1>
<div class="section" id="source-code">
<h2>Source code</h2>
<p>gzip-compressed source code: <a class="reference external" href="/ftp/python/2.5.5/Python-2.5.5.tgz">Python-2.5.5.tgz</a></p>
<p>bzip2-compressed source code: <a class="reference external" href="/ftp/python/2.5.5/Python-2.5.5.tar.bz2">Python-2.5.5.tar.bz2</a>,
the source archive.</p>
<p>The bzip2-compressed version is considerably smaller, so get that one if
your system has the <a class="reference external" href="http://www.bzip.org/">appropriate  tools</a> to deal
with it.</p>
<p>Unpack the archive with <tt class="docutils literal">tar <span class="pre">-zxvf</span> <span class="pre">Python-2.5.5.tgz</span></tt> (or
<tt class="docutils literal">bzcat <span class="pre">Python-2.5.5.tar.bz2</span> | tar <span class="pre">-xf</span> -</tt>).
Change to the Python-2.5.5 directory and run the &quot;./configure&quot;, &quot;make&quot;,
&quot;make install&quot; commands to compile and install Python. The source archive
is also suitable for Windows users who feel the need to build their
own version.</p>
</div>
</div>
<div class="section" id="what-s-new">
<h1>What's New?</h1>
<ul class="simple">
<li>See the <a class="reference external" href="../2.5/highlights">highlights</a> of the Python 2.5 release.</li>
<li>Andrew Kuchling's
<a class="reference external" href="http://www.python.org/doc/2.5/whatsnew/whatsnew25.html">What's New in Python 2.5</a>
describes the most visible changes since <a class="reference external" href="../2.4/">Python 2.4</a> in
more detail.</li>
<li>A detailed list of the changes in 2.5.5 can be found in
the <a class="reference external" href="NEWS.txt">release notes</a>, or the <tt class="docutils literal">Misc/NEWS</tt> file in the
source distribution.</li>
<li>For the full list of changes, you can poke around in
<a class="reference external" href="http://svn.python.org/view/python/branches/release25-maint/">Subversion</a>.</li>
</ul>
</div>
<div class="section" id="documentation">
<h1>Documentation</h1>
<p>The documentation has not been updated since the 2.5.4 release:</p>
<ul class="simple">
<li><a class="reference external" href="/doc/2.5.4/">Browse HTML on-line</a></li>
</ul>
</div>
<div class="section" id="files-md5-checksums-signatures-and-sizes">
<h1>Files, <a class="reference external" href="/download/releases/2.5.4/md5sum.py">MD5</a> checksums, signatures and sizes</h1>
<blockquote>
<p><tt class="docutils literal">abc02139ca38f4258e8e372f7da05c88</tt> <a class="reference external" href="/ftp/python/2.5.5/Python-2.5.5.tgz">Python-2.5.5.tgz</a>
(11606370 bytes, <a class="reference external" href="Python-2.5.5.tgz.asc">signature</a>)</p>
<p><tt class="docutils literal">1d00e2fb19418e486c30b850df625aa3</tt> <a class="reference external" href="/ftp/python/2.5.5/Python-2.5.5.tar.bz2">Python-2.5.5.tar.bz2</a>
(9822917 bytes, <a class="reference external" href="Python-2.5.5.tar.bz2.asc">signature</a>)</p>
</blockquote>
<p>The signatures above were generated with
<a class="reference external" href="http://www.gnupg.org">GnuPG</a> using release manager
Martin v. Löwis's <a class="reference external" href="/download#pubkeys">public key</a>
which has a key id of 7D9DC8D2.</p>
</div>
        <p><a href="http://hg.python.org/cpython/raw-file/v2.5.5/Misc/NEWS">Full Changelog</a></p>
        <header class="article-header">
            <h1 class="page-title">Files</h1>
        </header>
        <table>
          <thead>
            <tr>
              <th>Version</th>
              <th>Operating System</th>
              <th>Description</th>
              <th>MD5 Sum</th>
              <th>File Size</th>
              <th>GPG</th>
            </tr>
          </thead>
          <tbody>
              <tr>
                <td><a href="https://www.python.org/ftp/python/2.5.5/Python-2.5.5.tar.bz2">bzip2 compressed source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>1d00e2fb19418e486c30b850df625aa3</td>
                <td>9822917</td>
                <td><a href="https://www.python.org/ftp/python/2.5.5/Python-2.5.5.tar.bz2.asc">SIG</a></td>
              </tr>
              <tr>
                <td><a href="https://www.python.org/ftp/python/2.5.5/Python-2.5.5.tgz">Gzipped source tarball</a></td>
                <td>Source release</td>
                <td></td>
                <td>6953d49c4d2470d88d8577b4e5ed3ce2</td>
                <td>50155520</td>
                <td><a href="https://www.python.org/ftp/python/2.5.5/Python-2.5.5.tgz.asc">SIG</a></td>
              </tr>
          </tbody>
        </table>
    </article>
                </section>
            </div><!-- end .container -->
        </div><!-- end #content .content-wrapper -->
    </div><!-- end #touchnav-wrapper -->
</body>
</html>`
