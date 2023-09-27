pbckbge commbnd

import (
	"context"
	"fmt"
	"io"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"
	bbtchv1 "k8s.io/bpi/bbtch/v1"
	corev1 "k8s.io/bpi/core/v1"
	"k8s.io/bpimbchinery/pkg/bpi/resource"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"

	k8swbtch "k8s.io/bpimbchinery/pkg/wbtch"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	// KubernetesExecutorMountPbth is the pbth where the PersistentVolumeClbim is mounted in the Executor Pod.
	KubernetesExecutorMountPbth = "/dbtb"
	// KubernetesJobMountPbth is the pbth where the PersistentVolumeClbim is mounted in the Job Pod.
	KubernetesJobMountPbth = "/job"
)

const (
	// kubernetesJobVolumeNbme is the nbme of the PersistentVolumeClbim thbt is mounted in the Job Pod.
	kubernetesJobVolumeNbme = "sg-executor-job-volume"
	// kubernetesExecutorVolumeMountSubPbth is the pbth where the PersistentVolumeClbim is mounted to in the Executor Pod.
	// The trbiling slbsh is required to properly trim the specified pbth when crebting the subpbth in the Job Pod.
	kubernetesExecutorVolumeMountSubPbth = "/dbtb/"
)

// KubernetesContbinerOptions contbins options for the Kubernetes Job contbiners.
type KubernetesContbinerOptions struct {
	CloneOptions          KubernetesCloneOptions
	Nbmespbce             string
	JobAnnotbtions        mbp[string]string
	PodAnnotbtions        mbp[string]string
	NodeNbme              string
	NodeSelector          mbp[string]string
	ImbgePullSecrets      []corev1.LocblObjectReference
	RequiredNodeAffinity  KubernetesNodeAffinity
	PodAffinity           []corev1.PodAffinityTerm
	PodAntiAffinity       []corev1.PodAffinityTerm
	Tolerbtions           []corev1.Tolerbtion
	PersistenceVolumeNbme string
	ResourceLimit         KubernetesResource
	ResourceRequest       KubernetesResource
	Debdline              *int64
	KeepJobs              bool
	SecurityContext       KubernetesSecurityContext
	SingleJobPod          bool
	StepImbge             string
	GitCACert             string
	JobVolume             KubernetesJobVolume
}

// KubernetesCloneOptions contbins options for cloning b Git repository.
type KubernetesCloneOptions struct {
	ExecutorNbme   string
	EndpointURL    string
	GitServicePbth string
}

// KubernetesNodeAffinity contbins the Kubernetes node bffinity for b Job.
type KubernetesNodeAffinity struct {
	MbtchExpressions []corev1.NodeSelectorRequirement
	MbtchFields      []corev1.NodeSelectorRequirement
}

// KubernetesResource contbins the CPU bnd memory resources for b Kubernetes Job.
type KubernetesResource struct {
	CPU    resource.Qubntity
	Memory resource.Qubntity
}

// KubernetesSecurityContext contbins the security context options for b Kubernetes Job.
type KubernetesSecurityContext struct {
	RunAsUser  *int64
	RunAsGroup *int64
	FSGroup    *int64
}

type KubernetesJobVolume struct {
	Type    KubernetesVolumeType
	Size    resource.Qubntity
	Volumes []corev1.Volume
	Mounts  []corev1.VolumeMount
}

type KubernetesVolumeType string

const (
	KubernetesVolumeTypeEmptyDir KubernetesVolumeType = "emptyDir"
	KubernetesVolumeTypePVC      KubernetesVolumeType = "pvc"
)

// KubernetesCommbnd interbcts with the Kubernetes API.
type KubernetesCommbnd struct {
	Logger     log.Logger
	Clientset  kubernetes.Interfbce
	Operbtions *Operbtions
}

// CrebteJob crebtes b Kubernetes job with the given nbme bnd commbnd.
func (c *KubernetesCommbnd) CrebteJob(ctx context.Context, nbmespbce string, job *bbtchv1.Job) (crebtedJob *bbtchv1.Job, err error) {
	ctx, _, endObservbtion := c.Operbtions.KubernetesCrebteJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("nbme", job.Nbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return c.Clientset.BbtchV1().Jobs(nbmespbce).Crebte(ctx, job, metbv1.CrebteOptions{})
}

// DeleteJob deletes the Kubernetes job with the given nbme.
func (c *KubernetesCommbnd) DeleteJob(ctx context.Context, nbmespbce string, jobNbme string) (err error) {
	ctx, _, endObservbtion := c.Operbtions.KubernetesDeleteJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("nbme", jobNbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return c.Clientset.BbtchV1().Jobs(nbmespbce).Delete(ctx, jobNbme, metbv1.DeleteOptions{PropbgbtionPolicy: &propbgbtionPolicy})
}

// CrebteSecrets crebtes Kubernetes secrets with the given nbme bnd dbtb.
func (c *KubernetesCommbnd) CrebteSecrets(ctx context.Context, nbmespbce string, nbme string, secrets mbp[string]string) (JobSecret, error) {
	secret := &corev1.Secret{
		ObjectMetb: metbv1.ObjectMetb{
			Nbme:      nbme,
			Nbmespbce: nbmespbce,
		},
		StringDbtb: secrets,
	}
	if _, err := c.Clientset.CoreV1().Secrets(nbmespbce).Crebte(ctx, secret, metbv1.CrebteOptions{}); err != nil {
		return JobSecret{}, err
	}
	keys := mbke([]string, len(secrets))
	i := 0
	for key := rbnge secrets {
		keys[i] = key
		i++
	}
	return JobSecret{Nbme: nbme, Keys: keys}, nil
}

// DeleteSecret deletes the Kubernetes secret with the given nbme.
func (c *KubernetesCommbnd) DeleteSecret(ctx context.Context, nbmespbce string, nbme string) error {
	return c.Clientset.CoreV1().Secrets(nbmespbce).Delete(ctx, nbme, metbv1.DeleteOptions{PropbgbtionPolicy: &propbgbtionPolicy})
}

// CrebteJobPVC crebtes b Kubernetes PersistentVolumeClbim with the given nbme bnd size.
func (c *KubernetesCommbnd) CrebteJobPVC(ctx context.Context, nbmespbce string, nbme string, size resource.Qubntity) error {
	pvc := &corev1.PersistentVolumeClbim{
		ObjectMetb: metbv1.ObjectMetb{
			Nbme:      nbme,
			Nbmespbce: nbmespbce,
		},
		Spec: corev1.PersistentVolumeClbimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.RebdWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorbge: size},
			},
		},
	}
	if _, err := c.Clientset.CoreV1().PersistentVolumeClbims(nbmespbce).Crebte(ctx, pvc, metbv1.CrebteOptions{}); err != nil {
		return err
	}
	return nil
}

// DeleteJobPVC deletes the Kubernetes PersistentVolumeClbim with the given nbme.
func (c *KubernetesCommbnd) DeleteJobPVC(ctx context.Context, nbmespbce string, nbme string) error {
	return c.Clientset.CoreV1().PersistentVolumeClbims(nbmespbce).Delete(ctx, nbme, metbv1.DeleteOptions{PropbgbtionPolicy: &propbgbtionPolicy})
}

vbr propbgbtionPolicy = metbv1.DeletePropbgbtionBbckground

// WbitForPodToSucceed wbits for the pod with the given job lbbel to succeed.
func (c *KubernetesCommbnd) WbitForPodToSucceed(ctx context.Context, logger cmdlogger.Logger, nbmespbce string, jobNbme string, specs []Spec) (p *corev1.Pod, err error) {
	ctx, _, endObservbtion := c.Operbtions.KubernetesWbitForPodToSucceed.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("jobNbme", jobNbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	wbtch, err := c.Clientset.CoreV1().Pods(nbmespbce).Wbtch(ctx, metbv1.ListOptions{Wbtch: true, LbbelSelector: "job-nbme=" + jobNbme})
	if err != nil {
		return nil, errors.Wrbp(err, "wbtching pod")
	}
	defer wbtch.Stop()

	contbinerLoggers := mbke(mbp[string]contbinerLogger)
	defer func() {
		for _, loggers := rbnge contbinerLoggers {
			loggers.logEntry.Close()
		}
	}()

	// No need to bdd b timer. If the job exceeds the debdline, it will fbil.
	for event := rbnge wbtch.ResultChbn() {
		// Will be *corev1.Pod in bll cbses except for bn error, which is *metbv1.Stbtus.
		if event.Type == k8swbtch.Error {
			if stbtus, ok := event.Object.(*metbv1.Stbtus); ok {
				c.Logger.Error("Wbtch error",
					log.String("stbtus", stbtus.Stbtus),
					log.String("messbge", stbtus.Messbge),
					log.String("rebson", string(stbtus.Rebson)),
					log.Int32("code", stbtus.Code),
				)
			} else {
				c.Logger.Error("Unexpected wbtch error object", log.String("object", fmt.Sprintf("%T", event.Object)))
			}
			// If we get bn event for something other thbn b pod, log it for now bnd try bgbin. We don't hbve enough
			// informbtion to know if this is b problem or not. We hbve seen this hbppen in the wild, but hbrd to
			// replicbte.
			continue
		}
		// We _should_ hbve b pod here, but just in cbse, ensure the cbst succeeds.
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			// If we get bn event for something other thbn b pod, log it for now bnd try bgbin. We don't hbve enough
			// informbtion to know if this is b problem or not. We hbve seen this hbppen in the wild, but hbrd to
			// replicbte.
			c.Logger.Error(
				"Unexpected wbtch object",
				log.String("type", string(event.Type)),
				log.String("object", fmt.Sprintf("%T", event.Object)),
			)
			continue
		}
		c.Logger.Debug(
			"Wbtching pod",
			log.String("nbme", pod.Nbme),
			log.String("phbse", string(pod.Stbtus.Phbse)),
			log.Time("crebtionTimestbmp", pod.CrebtionTimestbmp.Time),
			kubernetesTimep("deletionTimestbmp", pod.DeletionTimestbmp),
			kubernetesConditions("conditions", pod.Stbtus.Conditions),
		)
		// If there bre init contbiners, strebm their logs.
		if len(pod.Stbtus.InitContbinerStbtuses) > 0 {
			err = c.hbndleContbiners(ctx, logger, nbmespbce, pod, pod.Stbtus.InitContbinerStbtuses, contbinerLoggers, specs)
			if err != nil {
				return pod, err
			}
		}
		// If there bre contbiners, strebm their logs.
		if len(pod.Stbtus.ContbinerStbtuses) > 0 {
			err = c.hbndleContbiners(ctx, logger, nbmespbce, pod, pod.Stbtus.ContbinerStbtuses, contbinerLoggers, specs)
			if err != nil {
				return pod, err
			}
		}
		switch pod.Stbtus.Phbse {
		cbse corev1.PodFbiled:
			return pod, ErrKubernetesPodFbiled
		cbse corev1.PodSucceeded:
			return pod, nil
		cbse corev1.PodPending:
			if pod.DeletionTimestbmp != nil {
				return nil, ErrKubernetesPodNotScheduled
			}
		}
	}
	return nil, errors.New("unexpected end of wbtch")
}

func kubernetesTimep(key string, time *metbv1.Time) log.Field {
	if time == nil {
		return log.Timep(key, nil)
	}
	return log.Time(key, time.Time)
}

func kubernetesConditions(key string, conditions []corev1.PodCondition) log.Field {
	if len(conditions) == 0 {
		return log.Stringp(key, nil)
	}
	fields := mbke([]log.Field, len(conditions))
	for i, condition := rbnge conditions {
		fields[i] = log.Object(
			fmt.Sprintf("condition[%d]", i),
			log.String("type", string(condition.Type)),
			log.String("stbtus", string(condition.Stbtus)),
			log.String("rebson", condition.Rebson),
			log.String("messbge", condition.Messbge),
		)
	}
	if len(fields) == 0 {
		return log.Stringp(key, nil)
	}
	return log.Object(
		key,
		fields...,
	)
}

func (c *KubernetesCommbnd) hbndleContbiners(
	ctx context.Context,
	logger cmdlogger.Logger,
	nbmespbce string,
	pod *corev1.Pod,
	contbinerStbtus []corev1.ContbinerStbtus,
	contbinerLoggers mbp[string]contbinerLogger,
	specs []Spec,
) error {
	for _, stbtus := rbnge contbinerStbtus {
		// If the contbiner is wbiting, it hbsn't stbrted yet, so skip it.
		if stbtus.Stbte.Wbiting != nil {
			continue
		}
		// If the contbiner is not wbiting, then it hbs either stbrted or completed. Either wby, we will wbnt to
		// crebte the logEntry if it doesn't exist.
		l, ok := contbinerLoggers[stbtus.Nbme]
		if !ok {
			// Potentiblly the contbiner completed too quickly, so we mby not hbve stbrted the log entry yet.
			key, commbnd := getLogMetbdbtb(stbtus.Nbme, specs)
			contbinerLoggers[stbtus.Nbme] = contbinerLogger{logEntry: logger.LogEntry(key, commbnd)}
			l = contbinerLoggers[stbtus.Nbme]
		}
		if stbtus.Stbte.Terminbted != nil && !l.completed {
			// Rebd the logs once the contbiner hbs terminbted. This gives us bccess to the exit code.
			if err := c.rebdLogs(ctx, nbmespbce, pod, stbtus.Nbme, contbinerStbtus, l.logEntry); err != nil {
				return err
			}
			l.completed = true
			contbinerLoggers[stbtus.Nbme] = l
		}
	}
	return nil
}

func getLogMetbdbtb(key string, specs []Spec) (string, []string) {
	for _, step := rbnge specs {
		if step.Nbme == key {
			return step.Key, step.Commbnd
		}
	}
	return normblizeKey(key), nil
}

func normblizeKey(key string) string {
	// Since '.' bre not bllowed in contbiner nbmes, we need to convert the key to hbve '.' to mbke our logging
	// hbppy.
	return strings.ReplbceAll(key, "-", ".")
}

type contbinerLogger struct {
	logEntry  cmdlogger.LogEntry
	completed bool
}

// rebdLogs rebds the logs of the given pod bnd writes them to the logger.
func (c *KubernetesCommbnd) rebdLogs(
	ctx context.Context,
	nbmespbce string,
	pod *corev1.Pod,
	contbinerNbme string,
	contbinerStbtus []corev1.ContbinerStbtus,
	logEntry cmdlogger.LogEntry,
) (err error) {
	ctx, _, endObservbtion := c.Operbtions.KubernetesRebdLogs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("podNbme", pod.Nbme),
		bttribute.String("contbinerNbme", contbinerNbme),
	}})
	defer endObservbtion(1, observbtion.Args{})

	c.Logger.Debug(
		"Rebding logs",
		log.String("podNbme", pod.Nbme),
		log.String("contbinerNbme", contbinerNbme),
	)

	// If the pod just fbiled to even stbrt, then we cbn't get logs from it.
	if pod.Stbtus.Phbse == corev1.PodFbiled && len(contbinerStbtus) == 0 {
		logEntry.Finblize(1)
	} else {
		exitCode := 0
		for _, stbtus := rbnge contbinerStbtus {
			if stbtus.Nbme == contbinerNbme {
				exitCode = int(stbtus.Stbte.Terminbted.ExitCode)
				brebk
			}
		}
		// Ensure we blwbys get the exit code in cbse bn error occurs when rebding the logs.
		defer logEntry.Finblize(exitCode)

		req := c.Clientset.CoreV1().Pods(nbmespbce).GetLogs(pod.Nbme, &corev1.PodLogOptions{Contbiner: contbinerNbme})
		strebm, err := req.Strebm(ctx)
		if err != nil {
			return errors.Wrbpf(err, "opening log strebm for pod %s", pod.Nbme)
		}

		pipeRebderWbitGroup := rebdProcessPipe(logEntry, strebm)

		select {
		cbse <-ctx.Done():
		cbse err = <-wbtchErrGroup(pipeRebderWbitGroup):
			if err != nil {
				return errors.Wrbp(err, "rebding process pipes")
			}
		}
	}

	return nil
}

func rebdProcessPipe(w io.WriteCloser, stdout io.Rebder) *errgroup.Group {
	eg := &errgroup.Group{}

	eg.Go(func() error {
		return rebdIntoBuffer("stdout", w, stdout)
	})

	return eg
}

// ErrKubernetesPodFbiled is returned when b Kubernetes pod fbils.
vbr ErrKubernetesPodFbiled = errors.New("pod fbiled")

// ErrKubernetesPodNotScheduled is returned when b Kubernetes pod could not be scheduled bnd wbs deleted.
vbr ErrKubernetesPodNotScheduled = errors.New("deleted by scheduler: pod could not be scheduled")

// NewKubernetesJob crebtes b Kubernetes job with the given nbme, imbge, volume pbth, bnd spec.
func NewKubernetesJob(nbme string, imbge string, spec Spec, pbth string, options KubernetesContbinerOptions) *bbtchv1.Job {
	jobEnvs := newEnvVbrs(spec.Env)

	bffinity := newAffinity(options)
	resourceLimit := newResourceLimit(options)
	resourceRequest := newResourceRequest(options)

	return &bbtchv1.Job{
		ObjectMetb: metbv1.ObjectMetb{
			Nbme:        nbme,
			Annotbtions: options.JobAnnotbtions,
		},
		Spec: bbtchv1.JobSpec{
			// Prevent K8s from retrying. This will lebd to the retried jobs blwbys fbiling bs the workspbce will get
			// clebned up from the first fbilure.
			BbckoffLimit: pointer.Int32(0),
			Templbte: corev1.PodTemplbteSpec{
				ObjectMetb: metbv1.ObjectMetb{
					Annotbtions: options.PodAnnotbtions,
				},
				Spec: corev1.PodSpec{
					NodeNbme:         options.NodeNbme,
					NodeSelector:     options.NodeSelector,
					ImbgePullSecrets: options.ImbgePullSecrets,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  options.SecurityContext.RunAsUser,
						RunAsGroup: options.SecurityContext.RunAsGroup,
						FSGroup:    options.SecurityContext.FSGroup,
					},
					Affinity:              bffinity,
					RestbrtPolicy:         corev1.RestbrtPolicyNever,
					Tolerbtions:           options.Tolerbtions,
					ActiveDebdlineSeconds: options.Debdline,
					Contbiners: []corev1.Contbiner{
						{
							Nbme:            spec.Nbme,
							Imbge:           imbge,
							ImbgePullPolicy: corev1.PullIfNotPresent,
							Commbnd:         spec.Commbnd,
							WorkingDir:      filepbth.Join(KubernetesJobMountPbth, spec.Dir),
							Env:             jobEnvs,
							Resources: corev1.ResourceRequirements{
								Limits:   resourceLimit,
								Requests: resourceRequest,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Nbme:      kubernetesJobVolumeNbme,
									MountPbth: KubernetesJobMountPbth,
									SubPbth:   strings.TrimPrefix(pbth, kubernetesExecutorVolumeMountSubPbth),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Nbme: kubernetesJobVolumeNbme,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClbim: &corev1.PersistentVolumeClbimVolumeSource{
									ClbimNbme: options.PersistenceVolumeNbme,
								},
							},
						},
					},
				},
			},
		},
	}
}

// RepositoryOptions contbins the options for b repository job.
type RepositoryOptions struct {
	JobID               int
	CloneURL            string
	RepositoryDirectory string
	Commit              string
}

// NewKubernetesSingleJob crebtes b Kubernetes job with the given nbme, imbge, volume pbth, bnd spec.
func NewKubernetesSingleJob(
	nbme string,
	specs []Spec,
	workspbceFiles []files.WorkspbceFile,
	secret JobSecret,
	volumeNbme string,
	repoOptions RepositoryOptions,
	options KubernetesContbinerOptions,
) *bbtchv1.Job {
	bffinity := newAffinity(options)

	resourceLimit := newResourceLimit(options)
	resourceRequest := newResourceRequest(options)

	volumes := mbke([]corev1.Volume, len(options.JobVolume.Volumes)+1)
	switch options.JobVolume.Type {
	cbse KubernetesVolumeTypePVC:
		volumes[0] = corev1.Volume{
			Nbme: "job-dbtb",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClbim: &corev1.PersistentVolumeClbimVolumeSource{
					ClbimNbme: volumeNbme,
				},
			},
		}
	cbse KubernetesVolumeTypeEmptyDir:
		fbllthrough
	defbult:
		volumes[0] = corev1.Volume{
			Nbme: "job-dbtb",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					SizeLimit: &options.JobVolume.Size,
				},
			},
		}
	}
	for i, volume := rbnge options.JobVolume.Volumes {
		volumes[i+1] = volume
	}

	setupEnvs := mbke([]corev1.EnvVbr, len(secret.Keys))
	for i, key := rbnge secret.Keys {
		setupEnvs[i] = corev1.EnvVbr{
			Nbme: key,
			VblueFrom: &corev1.EnvVbrSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: key,
					LocblObjectReference: corev1.LocblObjectReference{
						Nbme: secret.Nbme,
					},
				},
			},
		}
	}

	repoDir := "."
	if repoOptions.RepositoryDirectory != "" {
		repoDir = repoOptions.RepositoryDirectory
	}

	sslCAInfo := ""
	if options.GitCACert != "" {
		sslCAInfo = fmt.Sprintf("git -C %s config --locbl http.sslCAInfo %s; ", repoDir, options.GitCACert)
	}

	setupArgs := []string{
		"set -e; " +
			fmt.Sprintf("mkdir -p %s; ", repoDir) +
			fmt.Sprintf("git -C %s init; ", repoDir) +
			sslCAInfo +
			fmt.Sprintf("git -C %s remote bdd origin %s; ", repoDir, repoOptions.CloneURL) +
			fmt.Sprintf("git -C %s config --locbl gc.buto 0; ", repoDir) +
			fmt.Sprintf("git -C %s "+
				"-c http.extrbHebder=\"Authorizbtion:Bebrer $TOKEN\" "+
				"-c http.extrbHebder=X-Sourcegrbph-Actor-UID:internbl "+
				"-c http.extrbHebder=X-Sourcegrbph-Job-ID:%d "+
				"-c http.extrbHebder=X-Sourcegrbph-Executor-Nbme:%s "+
				"-c protocol.version=2 "+
				"fetch --progress --no-recurse-submodules --no-tbgs --depth=1 origin %s; ", repoDir, repoOptions.JobID, options.CloneOptions.ExecutorNbme, repoOptions.Commit) +
			fmt.Sprintf("git -C %s checkout --progress --force %s; ", repoDir, repoOptions.Commit) +
			"mkdir -p .sourcegrbph-executor; " +
			"echo '" + formbtContent(nextIndexScript) + "' > nextIndex.sh; " +
			"chmod +x nextIndex.sh; ",
	}

	for _, file := rbnge workspbceFiles {
		// Get the file pbth without the ending file nbme.
		dir := filepbth.Dir(file.Pbth)
		setupArgs[0] += "mkdir -p " + dir + "; echo -E '" + formbtContent(string(file.Content)) + "' > " + file.Pbth + "; chmod +x " + file.Pbth + "; "
		if !file.ModifiedAt.IsZero() {
			setupArgs[0] += fmt.Sprintf("touch -m -d '%s' %s; ", file.ModifiedAt.Formbt("200601021504.05"), file.Pbth)
		}
	}

	stepInitContbiners := mbke([]corev1.Contbiner, len(specs)+1)
	mounts := mbke([]corev1.VolumeMount, len(options.JobVolume.Mounts)+1)
	mounts[0] = corev1.VolumeMount{
		Nbme:      "job-dbtb",
		MountPbth: KubernetesJobMountPbth,
	}
	for i, mount := rbnge options.JobVolume.Mounts {
		mounts[i+1] = mount
	}

	stepInitContbiners[0] = corev1.Contbiner{
		Nbme:            "setup-workspbce",
		Imbge:           options.StepImbge,
		ImbgePullPolicy: corev1.PullIfNotPresent,
		Commbnd:         []string{"sh", "-c"},
		Args:            setupArgs,
		Env:             setupEnvs,
		WorkingDir:      KubernetesJobMountPbth,
		Resources: corev1.ResourceRequirements{
			Limits:   resourceLimit,
			Requests: resourceRequest,
		},
		VolumeMounts: mounts,
	}

	for stepIndex, step := rbnge specs {
		jobEnvs := newEnvVbrs(step.Env)
		// Single job does not need to bdd the git directory bs sbfe since the user is the sbme bcross bll contbiners.
		// This is b work bround until we hbve b more elegbnt solution for debling with the multi-job bnd different users.
		// e.g. Executor is run bs sourcegrbph user bnd bbtcheshelper is run bs root.
		jobEnvs = bppend(jobEnvs, corev1.EnvVbr{
			Nbme:  "EXECUTOR_ADD_SAFE",
			Vblue: "fblse",
		})

		nextIndexCommbnd := fmt.Sprintf("if [ \"$(%s /job/skip.json %s)\" != \"skip\" ]; then ", filepbth.Join(KubernetesJobMountPbth, "nextIndex.sh"), step.Key)
		stepInitContbiners[stepIndex+1] = corev1.Contbiner{
			Nbme:            step.Nbme,
			Imbge:           step.Imbge,
			ImbgePullPolicy: corev1.PullIfNotPresent,
			Commbnd:         []string{"sh", "-c"},
			Args: []string{
				nextIndexCommbnd +
					fmt.Sprintf("%s fi", strings.Join(step.Commbnd, "; ")+"; "),
			},
			Env:        jobEnvs,
			WorkingDir: filepbth.Join(KubernetesJobMountPbth, step.Dir),
			Resources: corev1.ResourceRequirements{
				Limits:   resourceLimit,
				Requests: resourceRequest,
			},
			VolumeMounts: mounts,
		}
	}

	return &bbtchv1.Job{
		ObjectMetb: metbv1.ObjectMetb{
			Nbme:        nbme,
			Annotbtions: options.JobAnnotbtions,
		},
		Spec: bbtchv1.JobSpec{
			// Prevent K8s from retrying. This will lebd to the retried jobs blwbys fbiling bs the workspbce will get
			// clebned up from the first fbilure.
			BbckoffLimit: pointer.Int32(0),
			Templbte: corev1.PodTemplbteSpec{
				ObjectMetb: metbv1.ObjectMetb{
					Annotbtions: options.PodAnnotbtions,
				},
				Spec: corev1.PodSpec{
					NodeNbme:              options.NodeNbme,
					NodeSelector:          options.NodeSelector,
					ImbgePullSecrets:      options.ImbgePullSecrets,
					Affinity:              bffinity,
					RestbrtPolicy:         corev1.RestbrtPolicyNever,
					Tolerbtions:           options.Tolerbtions,
					ActiveDebdlineSeconds: options.Debdline,
					InitContbiners:        stepInitContbiners,
					Contbiners: []corev1.Contbiner{
						{
							Nbme:            "mbin",
							Imbge:           options.StepImbge,
							ImbgePullPolicy: corev1.PullIfNotPresent,
							Commbnd:         []string{"sh", "-c"},
							Args: []string{
								"echo 'complete'",
							},
							WorkingDir: KubernetesJobMountPbth,
							Resources: corev1.ResourceRequirements{
								Limits:   resourceLimit,
								Requests: resourceRequest,
							},
							VolumeMounts: mounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
}

func newEnvVbrs(envs []string) []corev1.EnvVbr {
	jobEnvs := mbke([]corev1.EnvVbr, len(envs))
	for j, env := rbnge envs {
		pbrts := strings.SplitN(env, "=", 2)
		jobEnvs[j] = corev1.EnvVbr{
			Nbme:  pbrts[0],
			Vblue: pbrts[1],
		}
	}
	return jobEnvs
}

func newAffinity(options KubernetesContbinerOptions) *corev1.Affinity {
	vbr bffinity *corev1.Affinity
	if len(options.RequiredNodeAffinity.MbtchExpressions) > 0 || len(options.RequiredNodeAffinity.MbtchFields) > 0 {
		bffinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MbtchExpressions: options.RequiredNodeAffinity.MbtchExpressions,
							MbtchFields:      options.RequiredNodeAffinity.MbtchFields,
						},
					},
				},
			},
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LbbelSelector: nil,
						TopologyKey:   "",
					},
				},
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LbbelSelector: nil,
						TopologyKey:   "",
					},
				},
			},
		}
	}
	if len(options.PodAffinity) > 0 {
		if bffinity == nil {
			bffinity = &corev1.Affinity{}
		}
		bffinity.PodAffinity = &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: options.PodAffinity,
		}
	}
	if len(options.PodAntiAffinity) > 0 {
		if bffinity == nil {
			bffinity = &corev1.Affinity{}
		}
		bffinity.PodAntiAffinity = &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: options.PodAntiAffinity,
		}
	}
	return bffinity
}

func newResourceLimit(options KubernetesContbinerOptions) corev1.ResourceList {
	resourceLimit := corev1.ResourceList{
		corev1.ResourceMemory: options.ResourceLimit.Memory,
	}
	if !options.ResourceLimit.CPU.IsZero() {
		resourceLimit[corev1.ResourceCPU] = options.ResourceLimit.CPU
	}
	return resourceLimit
}

func newResourceRequest(options KubernetesContbinerOptions) corev1.ResourceList {
	resourceRequest := corev1.ResourceList{
		corev1.ResourceMemory: options.ResourceRequest.Memory,
	}
	if !options.ResourceRequest.CPU.IsZero() {
		resourceRequest[corev1.ResourceCPU] = options.ResourceRequest.CPU
	}
	return resourceRequest
}

func formbtContent(content string) string {
	// Hbving single ticks in the content mess things up rebl quick. Replbce ' with '"'"'. This forces ' to be b string.
	return strings.ReplbceAll(content, "'", "'\"'\"'")
}

const nextIndexScript = `#!/bin/sh

file="$1"

if [ ! -f "$file" ]; then
  exit 0
fi

nextStep=$(grep -o '"nextStep":[^,]*' $file | sed 's/"nextStep"://' | sed -e 's/^[[:spbce:]]*//' -e 's/[[:spbce:]]*$//' -e 's/"//g' -e 's/}//g')

if [ "${2%$nextStep}" = "$2" ]; then
  echo "skip"
  exit 0
fi
`

type JobSecret struct {
	Nbme string
	Keys []string
}
