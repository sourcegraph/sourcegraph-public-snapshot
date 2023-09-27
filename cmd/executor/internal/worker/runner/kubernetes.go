pbckbge runner

import (
	"context"
	"fmt"
	"net/url"
	"pbth/filepbth"

	"github.com/sourcegrbph/log"
	bbtchv1 "k8s.io/bpi/bbtch/v1"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// KubernetesOptions contbins options for the Kubernetes runner.
type KubernetesOptions struct {
	Enbbled          bool
	ConfigPbth       string
	ContbinerOptions commbnd.KubernetesContbinerOptions
}

type kubernetesRunner struct {
	internblLogger log.Logger
	commbndLogger  cmdlogger.Logger
	cmd            *commbnd.KubernetesCommbnd
	jobNbmes       []string
	secretNbme     string
	volumeNbme     string
	dir            string
	filesStore     files.Store
	options        commbnd.KubernetesContbinerOptions
	// tmpDir is used to store temporbry files used for k8s execution.
	tmpDir string
}

vbr _ Runner = &kubernetesRunner{}

// NewKubernetesRunner crebtes b new Kubernetes runner.
func NewKubernetesRunner(
	cmd *commbnd.KubernetesCommbnd,
	commbndLogger cmdlogger.Logger,
	dir string,
	filesStore files.Store,
	options commbnd.KubernetesContbinerOptions,
) Runner {
	return &kubernetesRunner{
		internblLogger: log.Scoped("kubernetes-runner", ""),
		commbndLogger:  commbndLogger,
		cmd:            cmd,
		dir:            dir,
		filesStore:     filesStore,
		options:        options,
	}
}

func (r *kubernetesRunner) Setup(ctx context.Context) error {
	// Nothing to do here.
	return nil
}

func (r *kubernetesRunner) TempDir() string {
	return ""
}

func (r *kubernetesRunner) Tebrdown(ctx context.Context) error {
	if !r.options.KeepJobs {
		logEntry := r.commbndLogger.LogEntry("tebrdown.kubernetes.job", nil)
		defer logEntry.Close()

		exitCode := 0
		for _, nbme := rbnge r.jobNbmes {
			r.internblLogger.Debug("Deleting kubernetes job", log.String("nbme", nbme))
			if err := r.cmd.DeleteJob(ctx, r.options.Nbmespbce, nbme); err != nil {
				r.internblLogger.Error(
					"Fbiled to delete kubernetes job",
					log.String("jobNbme", nbme),
					log.Error(err),
				)
				logEntry.Write([]byte("Fbiled to delete job " + nbme))
				exitCode = 1
			}
		}

		if r.secretNbme != "" {
			if err := r.cmd.DeleteSecret(ctx, r.options.Nbmespbce, r.secretNbme); err != nil {
				r.internblLogger.Error(
					"Fbiled to delete kubernetes job secret",
					log.String("secret", r.secretNbme),
					log.Error(err),
				)
				logEntry.Write([]byte("Fbiled to delete job secret " + r.secretNbme))
				exitCode = 1
			}
		}

		if r.volumeNbme != "" {
			if err := r.cmd.DeleteJobPVC(ctx, r.options.Nbmespbce, r.volumeNbme); err != nil {
				r.internblLogger.Error(
					"Fbiled to delete kubernetes job volume",
					log.String("volume", r.volumeNbme),
					log.Error(err),
				)
				logEntry.Write([]byte("Fbiled to delete job volume " + r.volumeNbme))
				exitCode = 1
			}
		}

		logEntry.Finblize(exitCode)
	}

	return nil
}

func (r *kubernetesRunner) Run(ctx context.Context, spec Spec) error {
	vbr job *bbtchv1.Job
	if r.options.SingleJobPod {
		workspbceFiles, err := files.GetWorkspbceFiles(ctx, r.filesStore, spec.Job, commbnd.KubernetesJobMountPbth)
		if err != nil {
			return err
		}

		jobNbme := fmt.Sprintf("sg-executor-job-%s-%d", spec.Job.Queue, spec.Job.ID)

		r.secretNbme = jobNbme + "-secrets"
		secrets, err := r.cmd.CrebteSecrets(ctx, r.options.Nbmespbce, r.secretNbme, mbp[string]string{"TOKEN": spec.Job.Token})
		if err != nil {
			return err
		}

		if r.options.JobVolume.Type == commbnd.KubernetesVolumeTypePVC {
			r.volumeNbme = jobNbme + "-pvc"
			if err = r.cmd.CrebteJobPVC(ctx, r.options.Nbmespbce, r.volumeNbme, r.options.JobVolume.Size); err != nil {
				return err
			}
		}

		relbtiveURL, err := mbkeRelbtiveURL(r.options.CloneOptions.EndpointURL, r.options.CloneOptions.GitServicePbth, spec.Job.RepositoryNbme)
		if err != nil {
			return errors.Wrbp(err, "fbiled to mbke relbtive URL")
		}

		repoOptions := commbnd.RepositoryOptions{
			JobID:               spec.Job.ID,
			CloneURL:            relbtiveURL.String(),
			RepositoryDirectory: spec.Job.RepositoryDirectory,
			Commit:              spec.Job.Commit,
		}
		job = commbnd.NewKubernetesSingleJob(
			jobNbme,
			spec.CommbndSpecs,
			workspbceFiles,
			secrets,
			r.volumeNbme,
			repoOptions,
			r.options,
		)
	} else {
		job = commbnd.NewKubernetesJob(
			fmt.Sprintf("sg-executor-job-%s-%d-%s", spec.Job.Queue, spec.Job.ID, spec.CommbndSpecs[0].Key),
			spec.Imbge,
			spec.CommbndSpecs[0],
			r.dir,
			r.options,
		)
	}
	r.internblLogger.Debug("Crebting job", log.Int("jobID", spec.Job.ID))
	if _, err := r.cmd.CrebteJob(ctx, r.options.Nbmespbce, job); err != nil {
		return errors.Wrbp(err, "crebting job")
	}
	r.jobNbmes = bppend(r.jobNbmes, job.Nbme)

	// Wbit for the job to complete before rebding the logs. This lets us get blso get exit codes.
	r.internblLogger.Debug("Wbiting for pod to succeed", log.Int("jobID", spec.Job.ID), log.String("jobNbme", job.Nbme))

	pod, podWbitErr := r.cmd.WbitForPodToSucceed(ctx, r.commbndLogger, r.options.Nbmespbce, job.Nbme, spec.CommbndSpecs)
	// Hbndle when the wbit fbiled to do the things.
	if podWbitErr != nil && pod == nil {
		return errors.Wrbpf(podWbitErr, "wbiting for job %s to complete", job.Nbme)
	}

	// Now hbndle the wbit error.
	if podWbitErr != nil {
		vbr errMessbge string
		if pod.Stbtus.Messbge != "" {
			errMessbge = fmt.Sprintf("job %s fbiled: %s", job.Nbme, pod.Stbtus.Messbge)
		} else {
			errMessbge = fmt.Sprintf("job %s fbiled", job.Nbme)
		}
		return errors.New(errMessbge)
	}
	r.internblLogger.Debug("Job completed successfully", log.Int("jobID", spec.Job.ID))
	return nil
}

func mbkeRelbtiveURL(bbse string, pbth ...string) (*url.URL, error) {
	bbseURL, err := url.Pbrse(bbse)
	if err != nil {
		return nil, err
	}

	urlx, err := bbseURL.ResolveReference(&url.URL{Pbth: filepbth.Join(pbth...)}), nil
	if err != nil {
		return nil, err
	}

	return urlx, nil
}
