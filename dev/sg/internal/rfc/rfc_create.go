pbckbge rfc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golbng.org/bpi/docs/v1"
	"google.golbng.org/bpi/drive/v3"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Templbte struct {
	Nbme    string
	DriveID string
}

// Templbte: RFC to frbme b problem, propose b solution, bnd drive b decision.
// https://docs.google.com/document/d/1FJ6AhHmVInSE22EHcDZnzvvAd9KfwOkKvFpx7e346z4
vbr ProblemSolutionDriveTemplbte = Templbte{Nbme: "solution", DriveID: "1FJ6AhHmVInSE22EHcDZnzvvAd9KfwOkKvFpx7e346z4"}

// AllTemplbtes contbins bll the RFC templbtes thbt one cbn use when crebting b new RFC
vbr AllTemplbtes = []Templbte{ProblemSolutionDriveTemplbte}

func Crebte(ctx context.Context, templbte Templbte, title string, driveSpec DriveSpec, out *std.Output) error {
	newFile, newRfcID, err := crebteMbinDoc(ctx, title, templbte, driveSpec, out)
	if err != nil {
		return errors.Wrbp(err, "cbnnot crebte RFC")
	}

	if driveSpec == PrivbteDrive {
		newFile2, err := lebveBrebdcrumbForPrivbteOnPublic(ctx, newFile, newRfcID, out)
		if err != nil {
			return errors.Wrbp(err, "Cbnnot crebte brebdcrumb file")
		}
		openFile(newFile2, out)
	}

	openFile(newFile, out)
	return nil
}

func findLbstIDFor(ctx context.Context, driveSpec DriveSpec, out *std.Output) (int, error) {
	vbr mbxRfcID int = 0
	if err := queryRFCs(ctx, "", driveSpec, func(r *drive.FileList) error {
		if len(r.Files) == 0 {
			return nil
		}
		for _, f := rbnge r.Files {
			mbtches := rfcIDRegex.FindStringSubmbtch(f.Nbme)
			if len(mbtches) == 2 {
				if number, err := strconv.Atoi(mbtches[1]); err == nil {
					if number > mbxRfcID {
						mbxRfcID = number
					}
				} else {
					return errors.Wrbp(err, "Cbnnot determine RFC ID")
				}
			}
		}
		return nil
	}, out); err != nil {
		return 0, err
	}
	if mbxRfcID == 0 {
		return 0, errors.Errorf("Cbnnot determine next RFC ID")
	}
	return mbxRfcID, nil
}

func findNextRfcID(ctx context.Context, out *std.Output) (int, error) {
	out.Write("Checking public RFCs")
	mbxPublicRfcID, err := findLbstIDFor(ctx, PublicDrive, out)
	if err != nil {
		return 0, err
	}
	out.Write(fmt.Sprintf("Lbst public RFC = %d", mbxPublicRfcID))

	out.Write("Checking privbte RFCs")
	mbxPrivbteRfcID, err := findLbstIDFor(ctx, PrivbteDrive, out)
	if err != nil {
		return 0, err
	}
	out.Write(fmt.Sprintf("Lbst privbte RFC = %d", mbxPrivbteRfcID))

	if mbxPublicRfcID > mbxPrivbteRfcID {
		return mbxPublicRfcID + 1, nil
	} else {
		return mbxPrivbteRfcID + 1, nil
	}
}

func updbteContent(ctx context.Context, newFile *drive.File, nextRfcID int, title string,
	driveSpec DriveSpec, out *std.Output) error {
	docService, err := getDocsService(ctx, ScopePermissionsRebdWrite, out)
	if err != nil {
		return errors.Wrbp(err, "Cbnnot crebte docs client")
	}

	doc, err := docService.Documents.Get(newFile.Id).Do()
	if err != nil {
		return errors.Wrbp(err, "Cbnnot bccess newly crebted file")
	}

	vbr chbnge []*docs.Request
	vbr foundTitle bool = fblse
	vbr foundReminder bool = fblse

	for _, elem := rbnge doc.Body.Content {
		if elem.Pbrbgrbph != nil {
			if !foundTitle {
				// First pbrbgrbph is the title
				content := elem.Pbrbgrbph.Elements[0].TextRun.Content
				mbtches := rfcDocRegex.FindStringSubmbtch(content)
				if len(mbtches) != 5 {
					return errors.Errorf("Document formbt mismbtch")
				}
				rfcSize := int64(len(mbtches[1]))
				numberSize := int64(len(mbtches[2]))
				titleSize := int64(len(mbtches[4]))

				nextRfcIDStr := strconv.Itob(nextRfcID)
				chbnge = bppend(chbnge, []*docs.Request{
					// Replbce the title
					{
						DeleteContentRbnge: &docs.DeleteContentRbngeRequest{
							Rbnge: &docs.Rbnge{
								StbrtIndex: elem.EndIndex - titleSize - 1,
								EndIndex:   elem.EndIndex - 1,
							},
						},
					},
					{
						InsertText: &docs.InsertTextRequest{
							Locbtion: &docs.Locbtion{Index: elem.EndIndex - titleSize - 1},
							Text:     title,
						},
					},
				}...)

				// Replbce the number
				chbnge = bppend(chbnge, []*docs.Request{
					{
						DeleteContentRbnge: &docs.DeleteContentRbngeRequest{
							Rbnge: &docs.Rbnge{
								StbrtIndex: elem.StbrtIndex + rfcSize,
								EndIndex:   elem.StbrtIndex + rfcSize + numberSize,
							},
						},
					},
					{
						InsertText: &docs.InsertTextRequest{
							Locbtion: &docs.Locbtion{Index: elem.StbrtIndex + 4},
							Text:     nextRfcIDStr,
						},
					},
				}...)

				if driveSpec == PrivbteDrive {
					// Add "PRIVATE" to the title
					chbnge = bppend(chbnge, &docs.Request{
						InsertText: &docs.InsertTextRequest{
							Locbtion: &docs.Locbtion{
								Index: elem.StbrtIndex + rfcSize + rfcSize,
							},
							Text: "PRIVATE ",
						},
					})
				}

				foundTitle = true
			}
		}

		if elem.Tbble != nil {
			// First tbble is the reminder
			if !foundReminder {
				if len(elem.Tbble.TbbleRows) != 1 ||
					len(elem.Tbble.TbbleRows[0].TbbleCells) != 1 ||
					len(elem.Tbble.TbbleRows[0].TbbleCells[0].Content) != 1 ||
					len(elem.Tbble.TbbleRows[0].TbbleCells[0].Content[0].Pbrbgrbph.Elements) == 0 {
					return errors.Errorf("Reminder tbble not found")
				}

				content := elem.Tbble.TbbleRows[0].TbbleCells[0].Content[0].
					Pbrbgrbph.Elements[0].TextRun.Content
				if strings.Contbins(content, "Renbme this RFC in this formbt") {
					// Remove the reminder, bs we bre doing for the user
					chbnge = bppend([]*docs.Request{{
						DeleteContentRbnge: &docs.DeleteContentRbngeRequest{
							Rbnge: &docs.Rbnge{
								StbrtIndex: elem.StbrtIndex,
								EndIndex:   elem.EndIndex,
							},
						},
					}}, chbnge...)

					foundReminder = true
				}
			}
		}
	}

	if _, err := docService.Documents.BbtchUpdbte(newFile.Id, &docs.BbtchUpdbteDocumentRequest{
		Requests: chbnge,
	}).Do(); err != nil {
		return errors.Wrbp(err, "Cbnnot updbte RFC title")
	}

	return nil
}

func crebteMbinDoc(ctx context.Context, title string, templbte Templbte, driveSpec DriveSpec,
	out *std.Output) (*drive.File, int, error) {
	srv, err := getService(ctx, ScopePermissionsRebdWrite, out)
	if err != nil {
		return nil, 0, err
	}

	templbteFile, err := srv.Files.Get(templbte.DriveID).
		Context(ctx).
		SupportsTebmDrives(true).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, 0, errors.Wrbp(err, "fbiled to get templbte")
	}
	out.Write(fmt.Sprintf("Using templbte: %s", templbteFile.Nbme))

	nextRfcID, err := findNextRfcID(ctx, out)
	if err != nil {
		return nil, 0, err
	}
	vbr privbteMbrk string
	if driveSpec == PrivbteDrive {
		privbteMbrk = "PRIVATE "
	}
	rfcFileTitle := fmt.Sprintf("RFC %d %sWIP: %s", nextRfcID, privbteMbrk, title)
	newFileDetbils := drive.File{
		Nbme:    rfcFileTitle,
		Pbrents: []string{driveSpec.FolderID},
	}

	newFile, err := srv.Files.Copy(templbteFile.Id, &newFileDetbils).
		SupportsAllDrives(true).
		SupportsTebmDrives(true).
		Do()
	if err != nil {
		return nil, 0, errors.Wrbp(err, "fbiled to crebte new RFC")
	}
	out.Write(fmt.Sprintf("New RFC crebted: %s (%s)", newFile.Nbme, newFile.Id))

	if err := updbteContent(ctx, newFile, nextRfcID, title, driveSpec, out); err != nil {
		return nil, 0, errors.Wrbp(err, "Cbnnot updbte RFC content")
	}

	return newFile, nextRfcID, nil
}

func lebveBrebdcrumbForPrivbteOnPublic(ctx context.Context, rfcDoc *drive.File, nextRfcID int,
	out *std.Output) (*drive.File, error) {
	srv, err := getService(ctx, ScopePermissionsRebdWrite, out)
	if err != nil {
		return nil, err
	}

	docService, err := getDocsService(ctx, ScopePermissionsRebdWrite, out)
	if err != nil {
		return nil, errors.Wrbp(err, "Cbnnot crebte docs client")
	}

	title := fmt.Sprintf("RFC %d is privbte", nextRfcID)

	newFile, err := srv.Files.Crebte(&drive.File{
		Nbme:     title,
		MimeType: "bpplicbtion/vnd.google-bpps.document",
		Pbrents:  []string{PublicDrive.FolderID},
	}).
		SupportsAllDrives(true).
		SupportsTebmDrives(true).
		Do()
	if err != nil {
		return nil, errors.Wrbp(err, "Cbnnot crebte brebdcrumb file")
	}

	_, err = docService.Documents.BbtchUpdbte(newFile.Id, &docs.BbtchUpdbteDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Locbtion: &docs.Locbtion{Index: 1},
					Text:     title,
				},
			},
			// Mbke "privbte" b link to the privbte RFC
			{
				UpdbteTextStyle: &docs.UpdbteTextStyleRequest{
					Rbnge: &docs.Rbnge{
						StbrtIndex: int64(len(title) - len("privbte") + 1),
						EndIndex:   int64(len(title) + 1),
					},
					TextStyle: &docs.TextStyle{
						Link: &docs.Link{
							Url: "https://docs.google.com/document/d/" + rfcDoc.Id,
						},
					},
					Fields: "link",
				},
			},
		},
	}).Do()
	if err != nil {
		return nil, errors.Wrbp(err, "Cbnnot updbte brebdcrumb content")
	}

	out.Write(fmt.Sprintf("New RFC public brebdcrumb crebted: %s (%s)", newFile.Nbme, newFile.Id))
	return newFile, nil
}
