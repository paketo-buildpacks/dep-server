package main_test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var entrypoint string

func TestEntrypoint(t *testing.T) {
	var Expect = NewWithT(t).Expect

	SetDefaultEventuallyTimeout(5 * time.Second)

	var err error
	entrypoint, err = gexec.Build("github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint")
	Expect(err).NotTo(HaveOccurred())

	spec.Run(t, "dispatch", func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually
		)

		context("", func() {
			var phpYMLFolder string

			it("sends a repository_dispatch webhook to a repo", func() {
				command := exec.Command(
					entrypoint,
					"--folder", phpYMLFolder,
				)
				buffer := gbytes.NewBuffer()

				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				//TODO: Check that the output matches the following structure:
				//{
				//  data: {
				//    php-8-yml:
				//    [
				//     {
				//        name: amqp
				//        version: 1.1.1
				//        md5: 12355
				//     } ,
				//     {
				//        name: redis
				//        version: 0.1.0
				//        d5: 5436346
				//     }
				//
				//    php-7-yml:
				//    [
				//    ]
				//  }
				//}

				Eventually(session).Should(gexec.Exit(0), func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) })
				//
				//Expect(buffer).To(gbytes.Say(`Success!`))

			})

			context("failure cases", func() {
				context("when the --folder flag is missing", func() {
					it("prints an error message and exits non-zero", func() {
						command := exec.Command(
							entrypoint,
						)
						buffer := gbytes.NewBuffer()

						session, err := gexec.Start(command, buffer, buffer)
						Expect(err).NotTo(HaveOccurred())

						Eventually(session).Should(gexec.Exit(1), func() string { return fmt.Sprintf("output:\n%s\n", buffer.Contents()) })

						Expect(buffer).To(gbytes.Say("the required flag `-f, --folder' was not specified"))
					})
				})
			})
		})
	}, spec.Report(report.Terminal{}), spec.Parallel())
}
