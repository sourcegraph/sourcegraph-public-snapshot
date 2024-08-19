// Copyright 2020-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by spdx-go-data. DO NOT EDIT.

package dataspdx

import "strings"

// LicenseInfo is SPDX license information.
//
// See https://spdx.org/licenses.
type LicenseInfo interface {
	// The SPDX identifier.
	//
	// Case-sensitive.
	ID() string
}

// GetLicenseInfo gets the LicenseInfo for the id.
//
// The parameter id is not case-sensitive.
func GetLicenseInfo(id string) (LicenseInfo, bool) {
	licenseInfo, ok := lowercaseIDToLicenseInfo[strings.ToLower(id)]
	return licenseInfo, ok
}

var lowercaseIDToLicenseInfo = map[string]*licenseInfo{
	"0bsd": {
		id: "0BSD",
	},
	"aal": {
		id: "AAL",
	},
	"abstyles": {
		id: "Abstyles",
	},
	"adobe-2006": {
		id: "Adobe-2006",
	},
	"adobe-glyph": {
		id: "Adobe-Glyph",
	},
	"adsl": {
		id: "ADSL",
	},
	"afl-1.1": {
		id: "AFL-1.1",
	},
	"afl-1.2": {
		id: "AFL-1.2",
	},
	"afl-2.0": {
		id: "AFL-2.0",
	},
	"afl-2.1": {
		id: "AFL-2.1",
	},
	"afl-3.0": {
		id: "AFL-3.0",
	},
	"afmparse": {
		id: "Afmparse",
	},
	"agpl-1.0": {
		id: "AGPL-1.0",
	},
	"agpl-1.0-only": {
		id: "AGPL-1.0-only",
	},
	"agpl-1.0-or-later": {
		id: "AGPL-1.0-or-later",
	},
	"agpl-3.0": {
		id: "AGPL-3.0",
	},
	"agpl-3.0-only": {
		id: "AGPL-3.0-only",
	},
	"agpl-3.0-or-later": {
		id: "AGPL-3.0-or-later",
	},
	"aladdin": {
		id: "Aladdin",
	},
	"amdplpa": {
		id: "AMDPLPA",
	},
	"aml": {
		id: "AML",
	},
	"ampas": {
		id: "AMPAS",
	},
	"antlr-pd": {
		id: "ANTLR-PD",
	},
	"antlr-pd-fallback": {
		id: "ANTLR-PD-fallback",
	},
	"apache-1.0": {
		id: "Apache-1.0",
	},
	"apache-1.1": {
		id: "Apache-1.1",
	},
	"apache-2.0": {
		id: "Apache-2.0",
	},
	"apafml": {
		id: "APAFML",
	},
	"apl-1.0": {
		id: "APL-1.0",
	},
	"app-s2p": {
		id: "App-s2p",
	},
	"apsl-1.0": {
		id: "APSL-1.0",
	},
	"apsl-1.1": {
		id: "APSL-1.1",
	},
	"apsl-1.2": {
		id: "APSL-1.2",
	},
	"apsl-2.0": {
		id: "APSL-2.0",
	},
	"arphic-1999": {
		id: "Arphic-1999",
	},
	"artistic-1.0": {
		id: "Artistic-1.0",
	},
	"artistic-1.0-cl8": {
		id: "Artistic-1.0-cl8",
	},
	"artistic-1.0-perl": {
		id: "Artistic-1.0-Perl",
	},
	"artistic-2.0": {
		id: "Artistic-2.0",
	},
	"baekmuk": {
		id: "Baekmuk",
	},
	"bahyph": {
		id: "Bahyph",
	},
	"barr": {
		id: "Barr",
	},
	"beerware": {
		id: "Beerware",
	},
	"bitstream-vera": {
		id: "Bitstream-Vera",
	},
	"bittorrent-1.0": {
		id: "BitTorrent-1.0",
	},
	"bittorrent-1.1": {
		id: "BitTorrent-1.1",
	},
	"blessing": {
		id: "blessing",
	},
	"blueoak-1.0.0": {
		id: "BlueOak-1.0.0",
	},
	"borceux": {
		id: "Borceux",
	},
	"bsd-1-clause": {
		id: "BSD-1-Clause",
	},
	"bsd-2-clause": {
		id: "BSD-2-Clause",
	},
	"bsd-2-clause-freebsd": {
		id: "BSD-2-Clause-FreeBSD",
	},
	"bsd-2-clause-netbsd": {
		id: "BSD-2-Clause-NetBSD",
	},
	"bsd-2-clause-patent": {
		id: "BSD-2-Clause-Patent",
	},
	"bsd-2-clause-views": {
		id: "BSD-2-Clause-Views",
	},
	"bsd-3-clause": {
		id: "BSD-3-Clause",
	},
	"bsd-3-clause-attribution": {
		id: "BSD-3-Clause-Attribution",
	},
	"bsd-3-clause-clear": {
		id: "BSD-3-Clause-Clear",
	},
	"bsd-3-clause-lbnl": {
		id: "BSD-3-Clause-LBNL",
	},
	"bsd-3-clause-modification": {
		id: "BSD-3-Clause-Modification",
	},
	"bsd-3-clause-no-military-license": {
		id: "BSD-3-Clause-No-Military-License",
	},
	"bsd-3-clause-no-nuclear-license": {
		id: "BSD-3-Clause-No-Nuclear-License",
	},
	"bsd-3-clause-no-nuclear-license-2014": {
		id: "BSD-3-Clause-No-Nuclear-License-2014",
	},
	"bsd-3-clause-no-nuclear-warranty": {
		id: "BSD-3-Clause-No-Nuclear-Warranty",
	},
	"bsd-3-clause-open-mpi": {
		id: "BSD-3-Clause-Open-MPI",
	},
	"bsd-4-clause": {
		id: "BSD-4-Clause",
	},
	"bsd-4-clause-shortened": {
		id: "BSD-4-Clause-Shortened",
	},
	"bsd-4-clause-uc": {
		id: "BSD-4-Clause-UC",
	},
	"bsd-protection": {
		id: "BSD-Protection",
	},
	"bsd-source-code": {
		id: "BSD-Source-Code",
	},
	"bsl-1.0": {
		id: "BSL-1.0",
	},
	"busl-1.1": {
		id: "BUSL-1.1",
	},
	"bzip2-1.0.5": {
		id: "bzip2-1.0.5",
	},
	"bzip2-1.0.6": {
		id: "bzip2-1.0.6",
	},
	"c-uda-1.0": {
		id: "C-UDA-1.0",
	},
	"cal-1.0": {
		id: "CAL-1.0",
	},
	"cal-1.0-combined-work-exception": {
		id: "CAL-1.0-Combined-Work-Exception",
	},
	"caldera": {
		id: "Caldera",
	},
	"catosl-1.1": {
		id: "CATOSL-1.1",
	},
	"cc-by-1.0": {
		id: "CC-BY-1.0",
	},
	"cc-by-2.0": {
		id: "CC-BY-2.0",
	},
	"cc-by-2.5": {
		id: "CC-BY-2.5",
	},
	"cc-by-2.5-au": {
		id: "CC-BY-2.5-AU",
	},
	"cc-by-3.0": {
		id: "CC-BY-3.0",
	},
	"cc-by-3.0-at": {
		id: "CC-BY-3.0-AT",
	},
	"cc-by-3.0-de": {
		id: "CC-BY-3.0-DE",
	},
	"cc-by-3.0-igo": {
		id: "CC-BY-3.0-IGO",
	},
	"cc-by-3.0-nl": {
		id: "CC-BY-3.0-NL",
	},
	"cc-by-3.0-us": {
		id: "CC-BY-3.0-US",
	},
	"cc-by-4.0": {
		id: "CC-BY-4.0",
	},
	"cc-by-nc-1.0": {
		id: "CC-BY-NC-1.0",
	},
	"cc-by-nc-2.0": {
		id: "CC-BY-NC-2.0",
	},
	"cc-by-nc-2.5": {
		id: "CC-BY-NC-2.5",
	},
	"cc-by-nc-3.0": {
		id: "CC-BY-NC-3.0",
	},
	"cc-by-nc-3.0-de": {
		id: "CC-BY-NC-3.0-DE",
	},
	"cc-by-nc-4.0": {
		id: "CC-BY-NC-4.0",
	},
	"cc-by-nc-nd-1.0": {
		id: "CC-BY-NC-ND-1.0",
	},
	"cc-by-nc-nd-2.0": {
		id: "CC-BY-NC-ND-2.0",
	},
	"cc-by-nc-nd-2.5": {
		id: "CC-BY-NC-ND-2.5",
	},
	"cc-by-nc-nd-3.0": {
		id: "CC-BY-NC-ND-3.0",
	},
	"cc-by-nc-nd-3.0-de": {
		id: "CC-BY-NC-ND-3.0-DE",
	},
	"cc-by-nc-nd-3.0-igo": {
		id: "CC-BY-NC-ND-3.0-IGO",
	},
	"cc-by-nc-nd-4.0": {
		id: "CC-BY-NC-ND-4.0",
	},
	"cc-by-nc-sa-1.0": {
		id: "CC-BY-NC-SA-1.0",
	},
	"cc-by-nc-sa-2.0": {
		id: "CC-BY-NC-SA-2.0",
	},
	"cc-by-nc-sa-2.0-fr": {
		id: "CC-BY-NC-SA-2.0-FR",
	},
	"cc-by-nc-sa-2.0-uk": {
		id: "CC-BY-NC-SA-2.0-UK",
	},
	"cc-by-nc-sa-2.5": {
		id: "CC-BY-NC-SA-2.5",
	},
	"cc-by-nc-sa-3.0": {
		id: "CC-BY-NC-SA-3.0",
	},
	"cc-by-nc-sa-3.0-de": {
		id: "CC-BY-NC-SA-3.0-DE",
	},
	"cc-by-nc-sa-3.0-igo": {
		id: "CC-BY-NC-SA-3.0-IGO",
	},
	"cc-by-nc-sa-4.0": {
		id: "CC-BY-NC-SA-4.0",
	},
	"cc-by-nd-1.0": {
		id: "CC-BY-ND-1.0",
	},
	"cc-by-nd-2.0": {
		id: "CC-BY-ND-2.0",
	},
	"cc-by-nd-2.5": {
		id: "CC-BY-ND-2.5",
	},
	"cc-by-nd-3.0": {
		id: "CC-BY-ND-3.0",
	},
	"cc-by-nd-3.0-de": {
		id: "CC-BY-ND-3.0-DE",
	},
	"cc-by-nd-4.0": {
		id: "CC-BY-ND-4.0",
	},
	"cc-by-sa-1.0": {
		id: "CC-BY-SA-1.0",
	},
	"cc-by-sa-2.0": {
		id: "CC-BY-SA-2.0",
	},
	"cc-by-sa-2.0-uk": {
		id: "CC-BY-SA-2.0-UK",
	},
	"cc-by-sa-2.1-jp": {
		id: "CC-BY-SA-2.1-JP",
	},
	"cc-by-sa-2.5": {
		id: "CC-BY-SA-2.5",
	},
	"cc-by-sa-3.0": {
		id: "CC-BY-SA-3.0",
	},
	"cc-by-sa-3.0-at": {
		id: "CC-BY-SA-3.0-AT",
	},
	"cc-by-sa-3.0-de": {
		id: "CC-BY-SA-3.0-DE",
	},
	"cc-by-sa-4.0": {
		id: "CC-BY-SA-4.0",
	},
	"cc-pddc": {
		id: "CC-PDDC",
	},
	"cc0-1.0": {
		id: "CC0-1.0",
	},
	"cddl-1.0": {
		id: "CDDL-1.0",
	},
	"cddl-1.1": {
		id: "CDDL-1.1",
	},
	"cdl-1.0": {
		id: "CDL-1.0",
	},
	"cdla-permissive-1.0": {
		id: "CDLA-Permissive-1.0",
	},
	"cdla-permissive-2.0": {
		id: "CDLA-Permissive-2.0",
	},
	"cdla-sharing-1.0": {
		id: "CDLA-Sharing-1.0",
	},
	"cecill-1.0": {
		id: "CECILL-1.0",
	},
	"cecill-1.1": {
		id: "CECILL-1.1",
	},
	"cecill-2.0": {
		id: "CECILL-2.0",
	},
	"cecill-2.1": {
		id: "CECILL-2.1",
	},
	"cecill-b": {
		id: "CECILL-B",
	},
	"cecill-c": {
		id: "CECILL-C",
	},
	"cern-ohl-1.1": {
		id: "CERN-OHL-1.1",
	},
	"cern-ohl-1.2": {
		id: "CERN-OHL-1.2",
	},
	"cern-ohl-p-2.0": {
		id: "CERN-OHL-P-2.0",
	},
	"cern-ohl-s-2.0": {
		id: "CERN-OHL-S-2.0",
	},
	"cern-ohl-w-2.0": {
		id: "CERN-OHL-W-2.0",
	},
	"clartistic": {
		id: "ClArtistic",
	},
	"cnri-jython": {
		id: "CNRI-Jython",
	},
	"cnri-python": {
		id: "CNRI-Python",
	},
	"cnri-python-gpl-compatible": {
		id: "CNRI-Python-GPL-Compatible",
	},
	"coil-1.0": {
		id: "COIL-1.0",
	},
	"community-spec-1.0": {
		id: "Community-Spec-1.0",
	},
	"condor-1.1": {
		id: "Condor-1.1",
	},
	"copyleft-next-0.3.0": {
		id: "copyleft-next-0.3.0",
	},
	"copyleft-next-0.3.1": {
		id: "copyleft-next-0.3.1",
	},
	"cpal-1.0": {
		id: "CPAL-1.0",
	},
	"cpl-1.0": {
		id: "CPL-1.0",
	},
	"cpol-1.02": {
		id: "CPOL-1.02",
	},
	"crossword": {
		id: "Crossword",
	},
	"crystalstacker": {
		id: "CrystalStacker",
	},
	"cua-opl-1.0": {
		id: "CUA-OPL-1.0",
	},
	"cube": {
		id: "Cube",
	},
	"curl": {
		id: "curl",
	},
	"d-fsl-1.0": {
		id: "D-FSL-1.0",
	},
	"diffmark": {
		id: "diffmark",
	},
	"dl-de-by-2.0": {
		id: "DL-DE-BY-2.0",
	},
	"doc": {
		id: "DOC",
	},
	"dotseqn": {
		id: "Dotseqn",
	},
	"drl-1.0": {
		id: "DRL-1.0",
	},
	"dsdp": {
		id: "DSDP",
	},
	"dvipdfm": {
		id: "dvipdfm",
	},
	"ecl-1.0": {
		id: "ECL-1.0",
	},
	"ecl-2.0": {
		id: "ECL-2.0",
	},
	"ecos-2.0": {
		id: "eCos-2.0",
	},
	"efl-1.0": {
		id: "EFL-1.0",
	},
	"efl-2.0": {
		id: "EFL-2.0",
	},
	"egenix": {
		id: "eGenix",
	},
	"elastic-2.0": {
		id: "Elastic-2.0",
	},
	"entessa": {
		id: "Entessa",
	},
	"epics": {
		id: "EPICS",
	},
	"epl-1.0": {
		id: "EPL-1.0",
	},
	"epl-2.0": {
		id: "EPL-2.0",
	},
	"erlpl-1.1": {
		id: "ErlPL-1.1",
	},
	"etalab-2.0": {
		id: "etalab-2.0",
	},
	"eudatagrid": {
		id: "EUDatagrid",
	},
	"eupl-1.0": {
		id: "EUPL-1.0",
	},
	"eupl-1.1": {
		id: "EUPL-1.1",
	},
	"eupl-1.2": {
		id: "EUPL-1.2",
	},
	"eurosym": {
		id: "Eurosym",
	},
	"fair": {
		id: "Fair",
	},
	"fdk-aac": {
		id: "FDK-AAC",
	},
	"frameworx-1.0": {
		id: "Frameworx-1.0",
	},
	"freebsd-doc": {
		id: "FreeBSD-DOC",
	},
	"freeimage": {
		id: "FreeImage",
	},
	"fsfap": {
		id: "FSFAP",
	},
	"fsful": {
		id: "FSFUL",
	},
	"fsfullr": {
		id: "FSFULLR",
	},
	"ftl": {
		id: "FTL",
	},
	"gd": {
		id: "GD",
	},
	"gfdl-1.1": {
		id: "GFDL-1.1",
	},
	"gfdl-1.1-invariants-only": {
		id: "GFDL-1.1-invariants-only",
	},
	"gfdl-1.1-invariants-or-later": {
		id: "GFDL-1.1-invariants-or-later",
	},
	"gfdl-1.1-no-invariants-only": {
		id: "GFDL-1.1-no-invariants-only",
	},
	"gfdl-1.1-no-invariants-or-later": {
		id: "GFDL-1.1-no-invariants-or-later",
	},
	"gfdl-1.1-only": {
		id: "GFDL-1.1-only",
	},
	"gfdl-1.1-or-later": {
		id: "GFDL-1.1-or-later",
	},
	"gfdl-1.2": {
		id: "GFDL-1.2",
	},
	"gfdl-1.2-invariants-only": {
		id: "GFDL-1.2-invariants-only",
	},
	"gfdl-1.2-invariants-or-later": {
		id: "GFDL-1.2-invariants-or-later",
	},
	"gfdl-1.2-no-invariants-only": {
		id: "GFDL-1.2-no-invariants-only",
	},
	"gfdl-1.2-no-invariants-or-later": {
		id: "GFDL-1.2-no-invariants-or-later",
	},
	"gfdl-1.2-only": {
		id: "GFDL-1.2-only",
	},
	"gfdl-1.2-or-later": {
		id: "GFDL-1.2-or-later",
	},
	"gfdl-1.3": {
		id: "GFDL-1.3",
	},
	"gfdl-1.3-invariants-only": {
		id: "GFDL-1.3-invariants-only",
	},
	"gfdl-1.3-invariants-or-later": {
		id: "GFDL-1.3-invariants-or-later",
	},
	"gfdl-1.3-no-invariants-only": {
		id: "GFDL-1.3-no-invariants-only",
	},
	"gfdl-1.3-no-invariants-or-later": {
		id: "GFDL-1.3-no-invariants-or-later",
	},
	"gfdl-1.3-only": {
		id: "GFDL-1.3-only",
	},
	"gfdl-1.3-or-later": {
		id: "GFDL-1.3-or-later",
	},
	"giftware": {
		id: "Giftware",
	},
	"gl2ps": {
		id: "GL2PS",
	},
	"glide": {
		id: "Glide",
	},
	"glulxe": {
		id: "Glulxe",
	},
	"glwtpl": {
		id: "GLWTPL",
	},
	"gnuplot": {
		id: "gnuplot",
	},
	"gpl-1.0": {
		id: "GPL-1.0",
	},
	"gpl-1.0+": {
		id: "GPL-1.0+",
	},
	"gpl-1.0-only": {
		id: "GPL-1.0-only",
	},
	"gpl-1.0-or-later": {
		id: "GPL-1.0-or-later",
	},
	"gpl-2.0": {
		id: "GPL-2.0",
	},
	"gpl-2.0+": {
		id: "GPL-2.0+",
	},
	"gpl-2.0-only": {
		id: "GPL-2.0-only",
	},
	"gpl-2.0-or-later": {
		id: "GPL-2.0-or-later",
	},
	"gpl-2.0-with-autoconf-exception": {
		id: "GPL-2.0-with-autoconf-exception",
	},
	"gpl-2.0-with-bison-exception": {
		id: "GPL-2.0-with-bison-exception",
	},
	"gpl-2.0-with-classpath-exception": {
		id: "GPL-2.0-with-classpath-exception",
	},
	"gpl-2.0-with-font-exception": {
		id: "GPL-2.0-with-font-exception",
	},
	"gpl-2.0-with-gcc-exception": {
		id: "GPL-2.0-with-GCC-exception",
	},
	"gpl-3.0": {
		id: "GPL-3.0",
	},
	"gpl-3.0+": {
		id: "GPL-3.0+",
	},
	"gpl-3.0-only": {
		id: "GPL-3.0-only",
	},
	"gpl-3.0-or-later": {
		id: "GPL-3.0-or-later",
	},
	"gpl-3.0-with-autoconf-exception": {
		id: "GPL-3.0-with-autoconf-exception",
	},
	"gpl-3.0-with-gcc-exception": {
		id: "GPL-3.0-with-GCC-exception",
	},
	"gsoap-1.3b": {
		id: "gSOAP-1.3b",
	},
	"haskellreport": {
		id: "HaskellReport",
	},
	"hippocratic-2.1": {
		id: "Hippocratic-2.1",
	},
	"hpnd": {
		id: "HPND",
	},
	"hpnd-sell-variant": {
		id: "HPND-sell-variant",
	},
	"htmltidy": {
		id: "HTMLTIDY",
	},
	"ibm-pibs": {
		id: "IBM-pibs",
	},
	"icu": {
		id: "ICU",
	},
	"ijg": {
		id: "IJG",
	},
	"imagemagick": {
		id: "ImageMagick",
	},
	"imatix": {
		id: "iMatix",
	},
	"imlib2": {
		id: "Imlib2",
	},
	"info-zip": {
		id: "Info-ZIP",
	},
	"intel": {
		id: "Intel",
	},
	"intel-acpi": {
		id: "Intel-ACPI",
	},
	"interbase-1.0": {
		id: "Interbase-1.0",
	},
	"ipa": {
		id: "IPA",
	},
	"ipl-1.0": {
		id: "IPL-1.0",
	},
	"isc": {
		id: "ISC",
	},
	"jam": {
		id: "Jam",
	},
	"jasper-2.0": {
		id: "JasPer-2.0",
	},
	"jpnic": {
		id: "JPNIC",
	},
	"json": {
		id: "JSON",
	},
	"lal-1.2": {
		id: "LAL-1.2",
	},
	"lal-1.3": {
		id: "LAL-1.3",
	},
	"latex2e": {
		id: "Latex2e",
	},
	"leptonica": {
		id: "Leptonica",
	},
	"lgpl-2.0": {
		id: "LGPL-2.0",
	},
	"lgpl-2.0+": {
		id: "LGPL-2.0+",
	},
	"lgpl-2.0-only": {
		id: "LGPL-2.0-only",
	},
	"lgpl-2.0-or-later": {
		id: "LGPL-2.0-or-later",
	},
	"lgpl-2.1": {
		id: "LGPL-2.1",
	},
	"lgpl-2.1+": {
		id: "LGPL-2.1+",
	},
	"lgpl-2.1-only": {
		id: "LGPL-2.1-only",
	},
	"lgpl-2.1-or-later": {
		id: "LGPL-2.1-or-later",
	},
	"lgpl-3.0": {
		id: "LGPL-3.0",
	},
	"lgpl-3.0+": {
		id: "LGPL-3.0+",
	},
	"lgpl-3.0-only": {
		id: "LGPL-3.0-only",
	},
	"lgpl-3.0-or-later": {
		id: "LGPL-3.0-or-later",
	},
	"lgpllr": {
		id: "LGPLLR",
	},
	"libpng": {
		id: "Libpng",
	},
	"libpng-2.0": {
		id: "libpng-2.0",
	},
	"libselinux-1.0": {
		id: "libselinux-1.0",
	},
	"libtiff": {
		id: "libtiff",
	},
	"liliq-p-1.1": {
		id: "LiLiQ-P-1.1",
	},
	"liliq-r-1.1": {
		id: "LiLiQ-R-1.1",
	},
	"liliq-rplus-1.1": {
		id: "LiLiQ-Rplus-1.1",
	},
	"linux-man-pages-copyleft": {
		id: "Linux-man-pages-copyleft",
	},
	"linux-openib": {
		id: "Linux-OpenIB",
	},
	"lpl-1.0": {
		id: "LPL-1.0",
	},
	"lpl-1.02": {
		id: "LPL-1.02",
	},
	"lppl-1.0": {
		id: "LPPL-1.0",
	},
	"lppl-1.1": {
		id: "LPPL-1.1",
	},
	"lppl-1.2": {
		id: "LPPL-1.2",
	},
	"lppl-1.3a": {
		id: "LPPL-1.3a",
	},
	"lppl-1.3c": {
		id: "LPPL-1.3c",
	},
	"lzma-sdk-9.11-to-9.20": {
		id: "LZMA-SDK-9.11-to-9.20",
	},
	"lzma-sdk-9.22": {
		id: "LZMA-SDK-9.22",
	},
	"makeindex": {
		id: "MakeIndex",
	},
	"minpack": {
		id: "Minpack",
	},
	"miros": {
		id: "MirOS",
	},
	"mit": {
		id: "MIT",
	},
	"mit-0": {
		id: "MIT-0",
	},
	"mit-advertising": {
		id: "MIT-advertising",
	},
	"mit-cmu": {
		id: "MIT-CMU",
	},
	"mit-enna": {
		id: "MIT-enna",
	},
	"mit-feh": {
		id: "MIT-feh",
	},
	"mit-modern-variant": {
		id: "MIT-Modern-Variant",
	},
	"mit-open-group": {
		id: "MIT-open-group",
	},
	"mitnfa": {
		id: "MITNFA",
	},
	"motosoto": {
		id: "Motosoto",
	},
	"mpi-permissive": {
		id: "mpi-permissive",
	},
	"mpich2": {
		id: "mpich2",
	},
	"mpl-1.0": {
		id: "MPL-1.0",
	},
	"mpl-1.1": {
		id: "MPL-1.1",
	},
	"mpl-2.0": {
		id: "MPL-2.0",
	},
	"mpl-2.0-no-copyleft-exception": {
		id: "MPL-2.0-no-copyleft-exception",
	},
	"mplus": {
		id: "mplus",
	},
	"ms-lpl": {
		id: "MS-LPL",
	},
	"ms-pl": {
		id: "MS-PL",
	},
	"ms-rl": {
		id: "MS-RL",
	},
	"mtll": {
		id: "MTLL",
	},
	"mulanpsl-1.0": {
		id: "MulanPSL-1.0",
	},
	"mulanpsl-2.0": {
		id: "MulanPSL-2.0",
	},
	"multics": {
		id: "Multics",
	},
	"mup": {
		id: "Mup",
	},
	"naist-2003": {
		id: "NAIST-2003",
	},
	"nasa-1.3": {
		id: "NASA-1.3",
	},
	"naumen": {
		id: "Naumen",
	},
	"nbpl-1.0": {
		id: "NBPL-1.0",
	},
	"ncgl-uk-2.0": {
		id: "NCGL-UK-2.0",
	},
	"ncsa": {
		id: "NCSA",
	},
	"net-snmp": {
		id: "Net-SNMP",
	},
	"netcdf": {
		id: "NetCDF",
	},
	"newsletr": {
		id: "Newsletr",
	},
	"ngpl": {
		id: "NGPL",
	},
	"nicta-1.0": {
		id: "NICTA-1.0",
	},
	"nist-pd": {
		id: "NIST-PD",
	},
	"nist-pd-fallback": {
		id: "NIST-PD-fallback",
	},
	"nlod-1.0": {
		id: "NLOD-1.0",
	},
	"nlod-2.0": {
		id: "NLOD-2.0",
	},
	"nlpl": {
		id: "NLPL",
	},
	"nokia": {
		id: "Nokia",
	},
	"nosl": {
		id: "NOSL",
	},
	"noweb": {
		id: "Noweb",
	},
	"npl-1.0": {
		id: "NPL-1.0",
	},
	"npl-1.1": {
		id: "NPL-1.1",
	},
	"nposl-3.0": {
		id: "NPOSL-3.0",
	},
	"nrl": {
		id: "NRL",
	},
	"ntp": {
		id: "NTP",
	},
	"ntp-0": {
		id: "NTP-0",
	},
	"nunit": {
		id: "Nunit",
	},
	"o-uda-1.0": {
		id: "O-UDA-1.0",
	},
	"occt-pl": {
		id: "OCCT-PL",
	},
	"oclc-2.0": {
		id: "OCLC-2.0",
	},
	"odbl-1.0": {
		id: "ODbL-1.0",
	},
	"odc-by-1.0": {
		id: "ODC-By-1.0",
	},
	"ofl-1.0": {
		id: "OFL-1.0",
	},
	"ofl-1.0-no-rfn": {
		id: "OFL-1.0-no-RFN",
	},
	"ofl-1.0-rfn": {
		id: "OFL-1.0-RFN",
	},
	"ofl-1.1": {
		id: "OFL-1.1",
	},
	"ofl-1.1-no-rfn": {
		id: "OFL-1.1-no-RFN",
	},
	"ofl-1.1-rfn": {
		id: "OFL-1.1-RFN",
	},
	"ogc-1.0": {
		id: "OGC-1.0",
	},
	"ogdl-taiwan-1.0": {
		id: "OGDL-Taiwan-1.0",
	},
	"ogl-canada-2.0": {
		id: "OGL-Canada-2.0",
	},
	"ogl-uk-1.0": {
		id: "OGL-UK-1.0",
	},
	"ogl-uk-2.0": {
		id: "OGL-UK-2.0",
	},
	"ogl-uk-3.0": {
		id: "OGL-UK-3.0",
	},
	"ogtsl": {
		id: "OGTSL",
	},
	"oldap-1.1": {
		id: "OLDAP-1.1",
	},
	"oldap-1.2": {
		id: "OLDAP-1.2",
	},
	"oldap-1.3": {
		id: "OLDAP-1.3",
	},
	"oldap-1.4": {
		id: "OLDAP-1.4",
	},
	"oldap-2.0": {
		id: "OLDAP-2.0",
	},
	"oldap-2.0.1": {
		id: "OLDAP-2.0.1",
	},
	"oldap-2.1": {
		id: "OLDAP-2.1",
	},
	"oldap-2.2": {
		id: "OLDAP-2.2",
	},
	"oldap-2.2.1": {
		id: "OLDAP-2.2.1",
	},
	"oldap-2.2.2": {
		id: "OLDAP-2.2.2",
	},
	"oldap-2.3": {
		id: "OLDAP-2.3",
	},
	"oldap-2.4": {
		id: "OLDAP-2.4",
	},
	"oldap-2.5": {
		id: "OLDAP-2.5",
	},
	"oldap-2.6": {
		id: "OLDAP-2.6",
	},
	"oldap-2.7": {
		id: "OLDAP-2.7",
	},
	"oldap-2.8": {
		id: "OLDAP-2.8",
	},
	"oml": {
		id: "OML",
	},
	"openssl": {
		id: "OpenSSL",
	},
	"opl-1.0": {
		id: "OPL-1.0",
	},
	"opubl-1.0": {
		id: "OPUBL-1.0",
	},
	"oset-pl-2.1": {
		id: "OSET-PL-2.1",
	},
	"osl-1.0": {
		id: "OSL-1.0",
	},
	"osl-1.1": {
		id: "OSL-1.1",
	},
	"osl-2.0": {
		id: "OSL-2.0",
	},
	"osl-2.1": {
		id: "OSL-2.1",
	},
	"osl-3.0": {
		id: "OSL-3.0",
	},
	"parity-6.0.0": {
		id: "Parity-6.0.0",
	},
	"parity-7.0.0": {
		id: "Parity-7.0.0",
	},
	"pddl-1.0": {
		id: "PDDL-1.0",
	},
	"php-3.0": {
		id: "PHP-3.0",
	},
	"php-3.01": {
		id: "PHP-3.01",
	},
	"plexus": {
		id: "Plexus",
	},
	"polyform-noncommercial-1.0.0": {
		id: "PolyForm-Noncommercial-1.0.0",
	},
	"polyform-small-business-1.0.0": {
		id: "PolyForm-Small-Business-1.0.0",
	},
	"postgresql": {
		id: "PostgreSQL",
	},
	"psf-2.0": {
		id: "PSF-2.0",
	},
	"psfrag": {
		id: "psfrag",
	},
	"psutils": {
		id: "psutils",
	},
	"python-2.0": {
		id: "Python-2.0",
	},
	"python-2.0.1": {
		id: "Python-2.0.1",
	},
	"qhull": {
		id: "Qhull",
	},
	"qpl-1.0": {
		id: "QPL-1.0",
	},
	"rdisc": {
		id: "Rdisc",
	},
	"rhecos-1.1": {
		id: "RHeCos-1.1",
	},
	"rpl-1.1": {
		id: "RPL-1.1",
	},
	"rpl-1.5": {
		id: "RPL-1.5",
	},
	"rpsl-1.0": {
		id: "RPSL-1.0",
	},
	"rsa-md": {
		id: "RSA-MD",
	},
	"rscpl": {
		id: "RSCPL",
	},
	"ruby": {
		id: "Ruby",
	},
	"sax-pd": {
		id: "SAX-PD",
	},
	"saxpath": {
		id: "Saxpath",
	},
	"scea": {
		id: "SCEA",
	},
	"schemereport": {
		id: "SchemeReport",
	},
	"sendmail": {
		id: "Sendmail",
	},
	"sendmail-8.23": {
		id: "Sendmail-8.23",
	},
	"sgi-b-1.0": {
		id: "SGI-B-1.0",
	},
	"sgi-b-1.1": {
		id: "SGI-B-1.1",
	},
	"sgi-b-2.0": {
		id: "SGI-B-2.0",
	},
	"shl-0.5": {
		id: "SHL-0.5",
	},
	"shl-0.51": {
		id: "SHL-0.51",
	},
	"simpl-2.0": {
		id: "SimPL-2.0",
	},
	"sissl": {
		id: "SISSL",
	},
	"sissl-1.2": {
		id: "SISSL-1.2",
	},
	"sleepycat": {
		id: "Sleepycat",
	},
	"smlnj": {
		id: "SMLNJ",
	},
	"smppl": {
		id: "SMPPL",
	},
	"snia": {
		id: "SNIA",
	},
	"spencer-86": {
		id: "Spencer-86",
	},
	"spencer-94": {
		id: "Spencer-94",
	},
	"spencer-99": {
		id: "Spencer-99",
	},
	"spl-1.0": {
		id: "SPL-1.0",
	},
	"ssh-openssh": {
		id: "SSH-OpenSSH",
	},
	"ssh-short": {
		id: "SSH-short",
	},
	"sspl-1.0": {
		id: "SSPL-1.0",
	},
	"standardml-nj": {
		id: "StandardML-NJ",
	},
	"sugarcrm-1.1.3": {
		id: "SugarCRM-1.1.3",
	},
	"swl": {
		id: "SWL",
	},
	"tapr-ohl-1.0": {
		id: "TAPR-OHL-1.0",
	},
	"tcl": {
		id: "TCL",
	},
	"tcp-wrappers": {
		id: "TCP-wrappers",
	},
	"tmate": {
		id: "TMate",
	},
	"torque-1.1": {
		id: "TORQUE-1.1",
	},
	"tosl": {
		id: "TOSL",
	},
	"tu-berlin-1.0": {
		id: "TU-Berlin-1.0",
	},
	"tu-berlin-2.0": {
		id: "TU-Berlin-2.0",
	},
	"ucl-1.0": {
		id: "UCL-1.0",
	},
	"unicode-dfs-2015": {
		id: "Unicode-DFS-2015",
	},
	"unicode-dfs-2016": {
		id: "Unicode-DFS-2016",
	},
	"unicode-tou": {
		id: "Unicode-TOU",
	},
	"unlicense": {
		id: "Unlicense",
	},
	"upl-1.0": {
		id: "UPL-1.0",
	},
	"vim": {
		id: "Vim",
	},
	"vostrom": {
		id: "VOSTROM",
	},
	"vsl-1.0": {
		id: "VSL-1.0",
	},
	"w3c": {
		id: "W3C",
	},
	"w3c-19980720": {
		id: "W3C-19980720",
	},
	"w3c-20150513": {
		id: "W3C-20150513",
	},
	"watcom-1.0": {
		id: "Watcom-1.0",
	},
	"wsuipa": {
		id: "Wsuipa",
	},
	"wtfpl": {
		id: "WTFPL",
	},
	"wxwindows": {
		id: "wxWindows",
	},
	"x11": {
		id: "X11",
	},
	"x11-distribute-modifications-variant": {
		id: "X11-distribute-modifications-variant",
	},
	"xerox": {
		id: "Xerox",
	},
	"xfree86-1.1": {
		id: "XFree86-1.1",
	},
	"xinetd": {
		id: "xinetd",
	},
	"xnet": {
		id: "Xnet",
	},
	"xpp": {
		id: "xpp",
	},
	"xskat": {
		id: "XSkat",
	},
	"ypl-1.0": {
		id: "YPL-1.0",
	},
	"ypl-1.1": {
		id: "YPL-1.1",
	},
	"zed": {
		id: "Zed",
	},
	"zend-2.0": {
		id: "Zend-2.0",
	},
	"zimbra-1.3": {
		id: "Zimbra-1.3",
	},
	"zimbra-1.4": {
		id: "Zimbra-1.4",
	},
	"zlib": {
		id: "Zlib",
	},
	"zlib-acknowledgement": {
		id: "zlib-acknowledgement",
	},
	"zpl-1.1": {
		id: "ZPL-1.1",
	},
	"zpl-2.0": {
		id: "ZPL-2.0",
	},
	"zpl-2.1": {
		id: "ZPL-2.1",
	},
}

type licenseInfo struct {
	id string
}

func (l *licenseInfo) ID() string {
	return l.id
}
