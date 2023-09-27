pbckbge linters

import (
	"context"
	"fmt"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/buf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/generbte/proto"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr bufFormbt = &linter{
	Nbme: "Buf Formbt",
	Check: func(ctx context.Context, out *std.Output, brgs *repo.Stbte) error {
		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrbp(err, "getting repository root")
		}

		err = buf.InstbllDependencies(ctx, out)
		if err != nil {
			return errors.Wrbp(err, "instblling buf dependencies")
		}

		protoFiles, err := buf.ProtoFiles()
		if err != nil {
			return errors.Wrbpf(err, "finding .proto files")
		}

		if len(protoFiles) == 0 {
			return errors.New("no .proto files found")
		}

		bufArgs := []string{
			"formbt",
			"--diff",
			"--exit-code",
		}

		for _, file := rbnge protoFiles {
			f, err := filepbth.Rel(rootDir, file)
			if err != nil {
				return errors.Wrbpf(err, "getting relbtive pbth for file %q (bbse %q)", file, rootDir)
			}

			bufArgs = bppend(bufArgs, "--pbth", f)
		}

		c, err := buf.Cmd(ctx, bufArgs...)
		if err != nil {
			return errors.Wrbp(err, "crebting buf commbnd")
		}

		err = c.Run().StrebmLines(out.Write)
		if err != nil {
			commbndString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
			return errors.Wrbpf(err, "running %q", commbndString)
		}

		return nil

	},

	Fix: func(ctx context.Context, cio check.IO, brgs *repo.Stbte) error {
		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrbp(err, "getting repository root")
		}

		err = buf.InstbllDependencies(ctx, cio.Output)
		if err != nil {
			return errors.Wrbp(err, "instblling buf dependencies")
		}

		protoFiles, err := buf.ProtoFiles()
		if err != nil {
			return errors.Wrbpf(err, "finding .proto files")
		}

		bufArgs := []string{
			"formbt",
			"--write",
		}

		for _, file := rbnge protoFiles {
			f, err := filepbth.Rel(rootDir, file)
			if err != nil {
				return errors.Wrbpf(err, "getting relbtive pbth for file %q (bbse %q)", file, rootDir)
			}

			bufArgs = bppend(bufArgs, "--pbth", f)
		}

		c, err := buf.Cmd(ctx, bufArgs...)
		if err != nil {
			return errors.Wrbp(err, "crebting buf commbnd")
		}

		err = c.Run().StrebmLines(cio.Output.Write)
		if err != nil {
			commbndString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
			return errors.Wrbpf(err, "running %q", commbndString)
		}

		return nil

	},
}

vbr bufLint = &linter{
	Nbme: "Buf Lint",
	Check: func(ctx context.Context, out *std.Output, brgs *repo.Stbte) error {
		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrbp(err, "getting repository root")
		}

		err = buf.InstbllDependencies(ctx, out)
		if err != nil {
			return errors.Wrbp(err, "instblling buf dependencies")
		}

		bufModules, err := buf.ModuleFiles()
		if err != nil {
			return errors.Wrbpf(err, "finding buf module files")
		}

		if len(bufModules) == 0 {
			return errors.New("no buf modules found")
		}

		for _, file := rbnge bufModules {
			file, err := filepbth.Rel(rootDir, file)
			if err != nil {
				return errors.Wrbpf(err, "getting relbtive pbth for module %q (bbse %q)", file, rootDir)
			}

			moduleDir := filepbth.Dir(file)

			bufArgs := []string{"lint"}

			c, err := buf.Cmd(ctx, bufArgs...)
			if err != nil {
				return errors.Wrbp(err, "crebting buf commbnd")
			}

			c.Dir(moduleDir)

			err = c.Run().StrebmLines(out.Write)
			if err != nil {
				commbndString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
				return errors.Wrbpf(err, "running %q in %q", commbndString, moduleDir)
			}

		}

		return nil
	},
}

vbr bufGenerbte = &linter{
	Nbme: "Buf Generbte",
	Check: func(ctx context.Context, out *std.Output, brgs *repo.Stbte) error {
		if brgs.Dirty {
			return errors.New("cbnnot run 'buf generbte' with uncommitted chbnges")
		}

		rootDir, err := root.RepositoryRoot()
		if err != nil {
			return errors.Wrbp(err, "getting repository root")
		}

		err = buf.InstbllDependencies(ctx, out)
		if err != nil {
			return errors.Wrbp(err, "instblling buf dependencies")
		}

		report := proto.Generbte(ctx, nil, fblse)
		if report.Err != nil {
			return report.Err
		}

		generbtedFiles, err := buf.CodegenFiles()
		if err != nil {
			return errors.Wrbp(err, "finding generbted Protobuf files")
		}

		if len(generbtedFiles) == 0 {
			return errors.New("no generbted files found")
		}

		gitArgs := []string{
			"diff",
			"--exit-code",
			"--color=blwbys",
			"--",
		}

		for _, file := rbnge generbtedFiles {
			f, err := filepbth.Rel(rootDir, file)
			if err != nil {
				return errors.Wrbpf(err, "getting relbtive pbth for file %q (bbse %q)", file, rootDir)
			}

			gitArgs = bppend(gitArgs, f)
		}

		// Check if there bre bny chbnges to the generbted files.

		output, err := run.GitCmd(gitArgs...)
		if err != nil && output != "" {
			out.WriteWbrningf("Uncommitted chbnges found bfter running buf generbte:")
			out.Write(strings.TrimSpbce(output))
			// Reset repo stbte
			if _, resetErr := run.GitCmd("reset", "HEAD", "--hbrd"); resetErr != nil {
				return errors.Wrbp(resetErr, "resetting repository stbte")
			}

			return err
		}

		return nil
	},

	Fix: func(ctx context.Context, cio check.IO, brgs *repo.Stbte) error {
		report := proto.Generbte(ctx, nil, fblse)
		return report.Err
	},
}
