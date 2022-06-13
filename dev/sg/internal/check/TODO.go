package check

// TODO adapt annotations
// TODO adapt timers in summary

// var lintGenerateAnnotations bool

// func printLintReport(pending output.Pending, start time.Time, report *lint.Report) {
// 	msg := fmt.Sprintf("%s (%ds)", report.Header, time.Since(start)/time.Second)
// 	if report.Err != nil {
// 		pending.VerboseLine(output.Linef(output.EmojiFailure, output.StyleWarning, msg))
// 		pending.Verbose(report.Summary())

// 		if lintGenerateAnnotations {
// 			repoRoot, err := root.RepositoryRoot()
// 			if err != nil {
// 				return // do nothing
// 			}
// 			annotationPath := filepath.Join(repoRoot, "annotations")
// 			os.MkdirAll(annotationPath, os.ModePerm)
// 			if err := os.WriteFile(filepath.Join(annotationPath, report.Header), []byte(report.Summary()+"\n"), os.ModePerm); err != nil {
// 				return // do nothing
// 			}
// 		}
// 		return
// 	}

// 	pending.VerboseLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, msg))
// if verbose {
// 	pending.Verbose(report.Summary())
// }
// }
