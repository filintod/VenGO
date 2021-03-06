/*
   Copyright (C) 2014  Oscar Campos <oscar.campos@member.fsf.org>

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 2 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License along
   with this program; if not, write to the Free Software Foundation, Inc.,
   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

   See LICENSE file for more details.
*/

package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/DamnWidget/VenGO/utils"
)

// testable type
type T struct {
	vcs *vcsType `json:"vcs"`
}

// vcs type structure
type vcsType struct {
	name      string
	refCmd    string
	updateCmd string
	cloneCmd  func(string, string, bool) error
	schemeCmd func(string, bool) (string, error)
}

// Git
var gitVcs = &vcsType{
	name:      "git",
	refCmd:    "git rev-parse --verify HEAD",
	updateCmd: "git checkout {tag}",
	cloneCmd: func(repo, tag string, verbose bool) error {
		if err := utils.Exec(verbose, "git", "clone", repo); err != nil {
			return err
		}
		curr, err := os.Getwd()
		if err != nil {
			return err
		}
		os.Chdir(path.Base(repo))
		err = utils.Exec(verbose, "git", "checkout", tag)
		os.Chdir(curr)
		return err
	},
	schemeCmd: func(repo string, verbose bool) (string, error) {
		for _, scheme := range []string{"git", "https", "http", "git+ssh"} {
			tmp := fmt.Sprintf("%s://%s", scheme, repo)
			if err := utils.Exec(verbose, "git", "ls-remote", tmp); err == nil {
				return scheme, nil
			}
		}
		return "", errors.New("scheme not found!")
	},
}

// Mercurial
var mercurialVcs = &vcsType{
	name:      "hg",
	refCmd:    "hg --debug id -i",
	updateCmd: "hg update -r {tag}",
	cloneCmd: func(repo, tag string, verbose bool) error {
		return utils.Exec(verbose, "hg", "clone", "-r", tag, repo)
	},
	schemeCmd: func(repo string, verbose bool) (string, error) {
		for _, scheme := range []string{"https", "http", "ssh"} {
			tmp := fmt.Sprintf("%s://%s", scheme, repo)
			if err := utils.Exec(verbose, "hg", "identify", tmp); err == nil {
				return scheme, nil
			}
		}
		return "", errors.New("scheme not found")
	},
}

// Bazaar
var bazaarVcs = &vcsType{
	name:      "bzr",
	refCmd:    "bzr revno",
	updateCmd: "bzr update -r revno:{tag}",
	cloneCmd: func(branch, rev string, verbose bool) error {
		return utils.Exec(verbose, "bzr", "branch", branch, "-r", rev)
	},
	schemeCmd: func(repo string, verbose bool) (string, error) {
		for _, scheme := range []string{"https", "http", "bzr", "bzr+ssh"} {
			tmp := fmt.Sprintf("%s://%s", scheme, repo)
			if err := utils.Exec(verbose, "bzr", "info", tmp); err == nil {
				return scheme, nil
			}
		}
		return "", errors.New("scheme not found")
	},
}

// SubVersion
var svnVcs = &vcsType{
	name:      "svn",
	refCmd:    `svn info | grep "Revision" | awk '{print $2}'`,
	updateCmd: "svn up -r{tag}",
	cloneCmd: func(repo, rev string, verbose bool) error {
		return utils.Exec(verbose, "svn", "checkout", "-r", rev, repo)
	},
	schemeCmd: func(repo string, verbose bool) (string, error) {
		for _, scheme := range []string{"https", "http", "svn", "svn+ssh"} {
			tmp := fmt.Sprintf("%s://%s", scheme, repo)
			if err := utils.Exec(verbose, "svn", "info", tmp); err == nil {
				return scheme, nil
			}
		}
		return "", errors.New("scheme not found")
	},
}

// available vcs types
var vcsTypes = []*vcsType{
	gitVcs,
	mercurialVcs,
	bazaarVcs,
	svnVcs,
}

// enable Unmarshaling of vcsType type
func (vcs *vcsType) UnmarshalJSON(b []byte) (err error) {
	var s string

	if err = json.Unmarshal(b, &s); err == nil {
		switch s {
		case "git":
			*vcs = *gitVcs
		case "hg":
			*vcs = *mercurialVcs
		case "bzr":
			*vcs = *bazaarVcs
		case "svn":
			*vcs = *svnVcs
		default:
			return fmt.Errorf("%s is not a valid vcs type")
		}
	} else {
		return err
	}
	return
}

// enable Marshaling of vcsType type
func (vcs *vcsType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, vcs.name)), nil
}

// clone the repo in an scpecific revision, tag or commit
func (vcs *vcsType) Clone(repo, tag, root string, verbose bool) error {
	// detect the right scheme to use very like go get does
	if scheme, err := vcs.schemeCmd(repo, verbose); err != nil {
		return err
	} else {
		curr, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(curr, root), 0755); err != nil {
			return err
		}
		if err := os.Chdir(root); err != nil {
			return err
		}
		defer os.Chdir(curr)
		return vcs.cloneCmd(fmt.Sprintf("%s://%s", scheme, repo), tag, verbose)
	}
}
