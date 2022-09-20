package docs

import (
	"bytes"
	"fmt"
	"github.com/werf/werf/cmd/werf/build"
	"github.com/werf/werf/cmd/werf/bundle/export"
	"github.com/werf/werf/cmd/werf/bundle/publish"
	"github.com/werf/werf/cmd/werf/cleanup"
	"github.com/werf/werf/cmd/werf/converge"
	"github.com/werf/werf/cmd/werf/dismiss"
	export2 "github.com/werf/werf/cmd/werf/export"
	"github.com/werf/werf/cmd/werf/purge"
	"html"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"mvdan.cc/xurls"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/common/templates"
)

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("{{ header }} Options\n\n```shell\n")
		buf.WriteString(templates.FlagsUsages(flags))
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("{{ header }} Options inherited from parent commands\n\n```shell\n")
		buf.WriteString(templates.FlagsUsages(parentFlags))
		buf.WriteString("```\n\n")
	}
	return nil
}

func printEnvironments(buf *bytes.Buffer, cmd *cobra.Command) error {
	environments, ok := cmd.Annotations[common.CmdEnvAnno]
	if !ok {
		return nil
	}

	if environments != "" {
		buf.WriteString("{{ header }} Environments\n\n```shell\n")
		buf.WriteString(environments)
		buf.WriteString("\n```\n\n")
	}

	return nil
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	buf := new(bytes.Buffer)

	buf.WriteString(`{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
`)

	buf.WriteString(replaceLinks(getLongFromCommand(cmd)) + "\n\n")

	if cmd.Runnable() {
		buf.WriteString("{{ header }} Syntax\n\n")
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", templates.UsageLine(cmd)))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("{{ header }} Examples\n\n")
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.Example))
	}

	if err := printEnvironments(buf, cmd); err != nil {
		return err
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func replaceLinks(s string) string {
	links := xurls.Relaxed.FindAllString(s, -1)
	for _, link := range links {
		linkText := link
		for _, prefix := range []string{"werf.io/documentation", "https://werf.io/documentation"} {
			if strings.HasPrefix(link, prefix) {
				link = strings.TrimPrefix(link, prefix)
				link = fmt.Sprintf("{{ \"%s\" | true_relative_url }}", link)
				break
			}
		}

		isSkipLink := false
		trimmedLink := strings.TrimPrefix(link, "https://")
		trimmedLink = strings.TrimPrefix(trimmedLink, "http://")
		for _, prefix := range []string{"another.example.com", "example.com", "localhost"} {
			if strings.HasPrefix(trimmedLink, prefix) {
				isSkipLink = true
				break
			}
		}

		if !isSkipLink {
			s = strings.ReplaceAll(s, linkText, fmt.Sprintf("[%s](%s)", linkText, link))
		}
	}

	return s
}

func fullCommandFilesystemPath(cmd string) string {
	res := cmd
	res = strings.ReplaceAll(res, " ", "_")
	res = strings.ReplaceAll(res, "-", "_")
	return res
}

func GenCliPages(cmdGroups templates.CommandGroups, pagesDir string) error {
	for _, group := range cmdGroups {
		for _, cmd := range group.Commands {
			if cmd.Hidden {
				continue
			}

			if err := genCliPages(cmd, pagesDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func genCliPages(cmd *cobra.Command, pagesDir string) error {
	fullCommandName := fullCommandFilesystemPath(cmd.CommandPath())
	cmdPage := fmt.Sprintf(`---
title: %s
permalink: reference/cli/%s.html
---

{%% include /reference/cli/%s.md %%}
`, cmd.CommandPath(), fullCommandName, fullCommandName)

	path := filepath.Join(pagesDir, fmt.Sprintf("%s.md", fullCommandName))
	if err := ioutil.WriteFile(path, []byte(cmdPage), 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}

	for _, command := range cmd.Commands() {
		if cmd.Hidden {
			continue
		}

		if err := genCliPages(command, pagesDir); err != nil {
			return err
		}
	}

	return nil
}

func GenCliSidebar(cmdGroups templates.CommandGroups, sidebarPath string) error {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(`# This file is generated by "werf docs" command.
# DO NOT EDIT!

cli: &cli
  - title: Overview of command groups
    url: /reference/cli/overview.html
`)

	for _, group := range cmdGroups {
		indent := 1
		groupRecord := fmt.Sprintf(`
%[1]s- title: %[2]s
%[1]s  f:`, strings.Repeat("  ", indent), group.Message)

		_, err := buf.WriteString(groupRecord)
		if err != nil {
			return err
		}

		indent += 2
		for _, cmd := range group.Commands {
			if cmd.Hidden {
				continue
			}

			if err := genCliSidebar(cmd, indent, buf); err != nil {
				return err
			}
		}
	}

	if err := ioutil.WriteFile(sidebarPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", sidebarPath, err)
	}

	return nil
}

func genCliSidebar(cmd *cobra.Command, indent int, buf *bytes.Buffer) error {
	if len(cmd.Commands()) == 0 {
		fullCommandName := fullCommandFilesystemPath(cmd.CommandPath())

		commandRecord := fmt.Sprintf(`
%[1]s- title: %[2]s
%[1]s  url: /reference/cli/%[3]s.html
`, strings.Repeat("  ", indent), cmd.CommandPath(), fullCommandName)

		_, err := buf.WriteString(commandRecord)
		if err != nil {
			return err
		}
	} else {
		groupRecord := fmt.Sprintf(`
%[1]s- title: %[2]s
%[1]s  f:`, strings.Repeat("  ", indent), cmd.CommandPath())

		_, err := buf.WriteString(groupRecord)
		if err != nil {
			return err
		}

		indent += 2
		for _, command := range cmd.Commands() {
			if cmd.Hidden {
				continue
			}

			if err := genCliSidebar(command, indent, buf); err != nil {
				return err
			}
		}
	}

	return nil
}

func GenCliOverview(cmdGroups templates.CommandGroups, pagesDir string) error {
	indexPage := `---
title: Overview of command groups
permalink: reference/cli/overview.html
toc: false
---

`

	var doNewline bool
	for _, group := range cmdGroups {
		if doNewline {
			indexPage += "\n"
		}

		indexPage += fmt.Sprintf("%s:\n", group.Message)

		for _, cmd := range group.Commands {
			if cmd.Hidden {
				continue
			}

			var fullCommandName string
			if len(cmd.Commands()) == 0 {
				fullCommandName = fullCommandFilesystemPath(cmd.CommandPath())
			} else {
				fullCommandName = fullCommandFilesystemPath(cmd.Commands()[0].CommandPath())
			}

			indexPage += fmt.Sprintf(" - [werf %s]({{ \"/reference/cli/%s.html\" | true_relative_url }}) — {%% include /reference/cli/%s.short.md %%}.\n", cmd.Name(), fullCommandName, fullCommandName)
		}

		doNewline = true
	}

	path := filepath.Join(pagesDir, "overview.md")
	if err := ioutil.WriteFile(path, []byte(indexPage), 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}

	return nil
}

func GenCliPartials(cmd *cobra.Command, dir string) error {
	for _, c := range cmd.Commands() {
		if cmd.Hidden {
			continue
		}

		if err := GenCliPartials(c, dir); err != nil {
			return err
		}
	}

	if err := writeFullCommandMarkdownPartial(cmd, dir); err != nil {
		return fmt.Errorf("unable to write full command partial: %w", err)
	}

	if err := writeShortCommandMarkdownPartial(cmd, dir); err != nil {
		return fmt.Errorf("unable to write full command partial: %w", err)
	}

	return nil
}

func writeFullCommandMarkdownPartial(cmd *cobra.Command, dir string) error {
	fullCommandName := fullCommandFilesystemPath(cmd.CommandPath())
	basename := fullCommandName + ".md"

	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := GenMarkdownCustom(cmd, f); err != nil {
		return err
	}
	return nil
}

func writeShortCommandMarkdownPartial(cmd *cobra.Command, dir string) error {
	fullCommandName := fullCommandFilesystemPath(cmd.CommandPath())
	basename := fullCommandName + ".short.md"
	path := filepath.Join(dir, basename)

	var desc string
	if len(cmd.Short) > 0 {
		desc = strings.ToLower(cmd.Short[0:1])
		desc += cmd.Short[1:]
	}

	if err := ioutil.WriteFile(path, []byte(desc), 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}

	return nil
}

func getLongFromCommand(cmd *cobra.Command) string {
	switch cmd.Short {
	case "Build and push images, then deploy application into Kubernetes":
		return html.EscapeString(converge.GetConvergeDocs().LongMD)
	case "Delete application from Kubernetes":
		return html.EscapeString(dismiss.GetDismissDocs().LongMD)
	case "Export bundle":
		return html.EscapeString(export.GetBundleExportDocs().LongMD)
	case "Publish bundle":
		return html.EscapeString(publish.GetBundlePublishDocs().LongMD)
	case "Cleanup project images in the container registry":
		return html.EscapeString(cleanup.GetCleanupDocs().LongMD)
	case "Purge all project images in the container registry":
		return html.EscapeString(purge.GetPurgeDocs().LongMD)
	case "Build images":
		return html.EscapeString(build.GetBuildDocs().LongMD)
	case "Export images":
		return html.EscapeString(export2.GetExportDocs().LongMD)
	default:
		if len(cmd.Long) == 0 {
			return cmd.Short
		} else {
			return cmd.Long
		}
	}
}
