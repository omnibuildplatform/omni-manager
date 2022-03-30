package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"omni-manager/models"
	"omni-manager/util"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// @Summary StartBuild Job
// @Description start a image build job
// @Tags  meta Manager
// @Param	body		body 	models.ImageInputData	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v1/images/startBuild [post]
func StartBuild(c *gin.Context) {

	var imageInputData models.ImageInputData
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	var insertData models.ImageMeta
	insertData.UserName = c.Keys["nm"].(string)
	insertData.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))

	insertData.Arch = imageInputData.Arch
	insertData.Release = imageInputData.Release
	insertData.BuildType = imageInputData.BuildType
	if len(insertData.Release) == 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "Release not allowed empty ", nil))
		return
	}

	if insertData.UserId <= 0 {
		c.JSON(http.StatusForbidden, util.ExportData(util.CodeStatusClientError, " forbidden ", nil))
		return
	}
	//check package validate
	validate := false
	for _, arch := range util.GetConfig().BuildParam.Arch {
		if arch == insertData.Arch {
			validate = true
			break
		}
	}
	if !validate {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "arch not supported  ", util.GetConfig().BuildParam.Arch))
		return
	}
	validate = false //reset for buildtype
	for _, buildtype := range util.GetConfig().BuildParam.BuildType {
		if buildtype == insertData.BuildType {
			validate = true
			break
		}
	}
	if !validate {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "buildType not supported  ", util.GetConfig().BuildParam.BuildType))
		return
	}

	insertData.CustomPkg = strings.Join(imageInputData.CustomPkg, ",")
	controllerID := uuid.NewV4().String()
	var jobName = fmt.Sprintf(`omni-image-%s`, controllerID)
	var imageName = fmt.Sprintf(`openEuler-%s.iso`, controllerID)
	clientset, err := kubernetes.NewForConfig(models.GetK8sConfig())
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "kubernetes.NewForConfig Error:  ", err))
		return
	}

	//-------- make custom rpms config first

	totalPkgs := make(map[string][]string)
	totalPkgs["packages"] = append(util.GetConfig().DefaultPkgList.Packages, imageInputData.CustomPkg...)
	confYmalConentBytes, err := json.Marshal(totalPkgs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "json marshal total rpm names Errof: ", err))
		return
	}

	insertData.ConfigMapName = fmt.Sprintf("cmname%d", time.Now().UnixMicro())
	// cfgapp := apconfigcorev1.ConfigMap(insertData.ConfigMapName, util.GetConfig().K8sConfig.Namespace)
	tempdata := make(map[string]string)
	tempdata["working_dir"] = "/opt/omni-workspace"
	tempdata["debug"] = "True"
	tempdata["user_name"] = "root"
	tempdata["user_passwd"] = "openEuler"
	tempdata["installer_configs"] = "/etc/omni-imager/installer_assets/calamares-configs"
	tempdata["systemd_configs"] = "/etc/omni-imager/installer_assets/systemd-configs"
	tempdata["init_script"] = "/etc/omni-imager/init"
	tempdata["installer_script"] = "/etc/omni-imager/runinstaller"
	tempdata["repo_file"] = fmt.Sprintf("/etc/omni-imager/repos/%s.repo", insertData.Release)
	tempdataBytes, _ := yaml.Marshal(tempdata)
	// cfgapp.WithData(map[string]string{
	// 	"conf.yaml":      string(tempdataBytes),
	// 	"totalrpms.json": string(confYmalConentBytes),
	// }).WithAPIVersion("v1").WithKind("ConfigMap")

	configMapType := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ConfigMap",
	}
	var ownerReferenceController bool = true
	var BlockOwnerDeletion bool = false

	var configImage *v1.ConfigMap
	configImage = &v1.ConfigMap{
		TypeMeta: configMapType,
		ObjectMeta: metav1.ObjectMeta{
			Name:      insertData.ConfigMapName,
			Namespace: util.GetConfig().K8sConfig.Namespace,
		},
		Data: map[string]string{
			"conf.yaml":      string(tempdataBytes),
			"totalrpms.json": string(confYmalConentBytes),
		},
	}
	var cm *v1.ConfigMap
	cm, err = clientset.CoreV1().ConfigMaps(util.GetConfig().K8sConfig.Namespace).Create(context.TODO(), configImage, metav1.CreateOptions{
		TypeMeta: configImage.TypeMeta,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "create job image  Errof: ", err))
		return
	}
	cm.TypeMeta = configMapType

	//---------------------------------
	omniImager := `omni-imager --package-list /conf/totalrpms.json --config-file /conf/conf.yaml --build-type ` + insertData.BuildType + ` --output-file ` + imageName + ` && curl -vvv -Ffile=@/opt/omni-workspace/` + imageName + ` -Fproject=` + insertData.Release + `  -FfileType=image '` + util.GetConfig().K8sConfig.FfileType + `'`
	jobs := clientset.BatchV1().Jobs(util.GetConfig().K8sConfig.Namespace)
	var backOffLimit int32 = 0
	var tTLSecondsAfterFinished int32 = 1800
	var privileged bool = true

	jobYaml := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: util.GetConfig().K8sConfig.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         configMapType.APIVersion,
					Kind:               configMapType.Kind,
					Name:               cm.Name,
					Controller:         &ownerReferenceController,
					UID:                cm.UID,
					BlockOwnerDeletion: &BlockOwnerDeletion,
				},
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  jobName,
							Image: util.GetConfig().K8sConfig.Image,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Command: []string{
								"/bin/sh",
								"-c",
								omniImager,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "confyaml",
									MountPath: "/conf",
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{
							Name: "confyaml",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: insertData.ConfigMapName,
									},
								},
							},
						},
					},
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &tTLSecondsAfterFinished,
		},
	}

	//----------create job

	job, err := jobs.Create(context.TODO(), jobYaml, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "Create job Error", err))
		return

	}

	// tempbytes, _ := json.Marshal(job)
	// fmt.Println("--------------:", string(tempbytes))

	insertData.UserName = c.Keys["nm"].(string)
	insertData.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	insertData.Status = models.JOB_STATUS_RUNNING
	insertData.JobName = job.GetName()
	insertData.CreateTime = job.GetCreationTimestamp().Time
	insertData.DownloadUrl = fmt.Sprintf(util.GetConfig().BuildParam.DownloadIsoUrl, insertData.Release, time.Now().Format("2006-01-02"), imageName)
	jobDBID, err := models.AddImageMeta(&insertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, jobDBID, job.GetName(), util.GetConfig().WSConfig))
}

// @Summary QueryJobStatus
// @Description QueryJobStatus for given job name
// @Tags  meta Manager
// @Param	name		path 	string	true		"The name for job"
// @Param	id		query 	string	false		"The id for job in database. "
// @Param	ns		query 	string	false		"job namespace "
// @Accept json
// @Produce json
// @Router /v1/images/queryJobStatus/{name} [get]
func QueryJobStatus(c *gin.Context) {
	jobname := c.Param("name")
	if jobname == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job name must be fill:", nil))
		return
	}
	jobNamespace, _ := c.GetQuery("ns")
	if jobNamespace == "" {
		//if no special,then use config namespace
		jobNamespace = util.GetConfig().K8sConfig.Namespace
	}
	jobidStr, _ := c.GetQuery("id")
	// if given jobid . update job status in database
	jobid, _ := strconv.Atoi(jobidStr)
	var err error
	var imageData *models.ImageMeta
	if jobid > 0 {
		imageData, err = models.GetImageMetaById(jobid)
	}

	jobAPI := models.GetClientSet().BatchV1()
	var job *batchv1.Job
	job, err = jobAPI.Jobs(jobNamespace).Get(context.TODO(), jobname, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, " QueryJobStatus Error:", err))
		return
	}
	completions := job.Spec.Completions
	backoffLimit := job.Spec.BackoffLimit
	result := make(map[string]interface{})
	result["name"] = jobname
	result["startTime"] = job.Status.StartTime
	// check status
	if job.Status.Succeeded >= *completions {
		result["status"] = models.JOB_STATUS_SUCCEED
		result["completionTime"] = job.Status.CompletionTime
		if imageData != nil {
			result["url"] = imageData.DownloadUrl
			var updateJob models.ImageMeta
			updateJob.JobName = jobname
			updateJob.Id = jobid
			updateJob.Status = models.JOB_STATUS_SUCCEED
			models.UpdateJobStatus(&updateJob)
		}

		// go func() {
		//   delete this ConfigMap
		// 	clientset.CoreV1().ConfigMaps(util.GetConfig().K8sConfig.Namespace).Delete(context.TODO(), insertData.ConfigMapName, *metav1.NewDeleteOptions(0))
		// }()
		//no export if success
		job = nil
	} else if job.Status.Failed >= *backoffLimit {
		result["status"] = models.JOB_STATUS_FAILED
		result["error"] = job.Status.String()
		result["completionTime"] = job.Status.CompletionTime
		if imageData != nil {
			var updateJob models.ImageMeta
			updateJob.JobName = jobname
			updateJob.Id = jobid
			updateJob.Status = models.JOB_STATUS_FAILED
			models.UpdateJobStatus(&updateJob)
		}
	} else if job.Status.Succeeded == 0 || job.Status.Failed == 0 {
		result["status"] = models.JOB_STATUS_RUNNING
	}

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result, job))
}

// @Summary QueryJobLogs
// @Description QueryJobLogs for given job name
// @Tags  meta Manager
// @Param	name		path 	string	true		"The name for job"
// @Accept json
// @Produce json
// @Router /v1/images/queryJobLogs/{name} [get]
func QueryJobLogs(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job name must be fill:", nil))
		return
	}
	listopt := metav1.ListOptions{}
	listopt.LabelSelector = "job-name=" + name
	pods, err := models.GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).List(context.TODO(), listopt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "", err))
		return
	}
	buf := new(bytes.Buffer)
	for _, pod := range pods.Items {
		buf.WriteString(fmt.Sprintf("------------------- pod.name:%s \n---------", pod.Name))
		req := models.GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(context.TODO())
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "", err))
			return
		}
		defer podLogs.Close()
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "", err))
			return
		}
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", buf.String()))
}

// @Summary get
// @Description get single one
// @Tags  meta Manager
// @Param	id		path 	string	true		"The key for staticblock"
// @Accept json
// @Produce json
// @Router /v1/images/get/{id} [get]
func Read(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if id <= 0 || err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "id must be int type", err))
		return
	}
	v, err := models.GetImageMetaById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, id, v))
}

// @Summary GetBaseData param
// @Description get architecture, release Version, output Format ,and default package name list
// @Tags  meta Manager
// @Accept json
// @Produce json
// @Router /v1/images/param/getBaseData/ [get]
func GetBaseData(c *gin.Context) {

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal,
		"ok", util.GetConfig().BuildParam, util.GetConfig().DefaultPkgList, models.CustomSigList))
}

// @Summary GetCustomePkgList param
// @Description get custom package name list
// @Tags  meta Manager
// @Param	arch		query 	string	true		" arch ,e g:x86_64"
// @Param	release		query 	string	true		"release  "
// @Param	sig		query 	string	true		"custom group  "
// @Accept json
// @Produce json
// @Router /v1/images/param/getCustomePkgList/ [get]
func GetCustomePkgList(c *gin.Context) {

	arch := c.Query("arch")
	release := c.Query("release")
	sig := c.Query("sig")

	customlist, err := models.GetCustomePkgList(release, arch, sig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", customlist))
}

// @Summary query multi datas
// @Description use param to query multi datas
// @Tags  meta Manager
// @Param	project_name		query 	string	true		"project name"
// @Param	pkg_name		query 	string	true		"package name"
// @Accept json
// @Produce json
// @Router /v1/images/query [get]
func Query(c *gin.Context) {
	//...... emplty . wait for query param
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, c.Query("project_name"), c.Query("pkg_name")))
}

// @Summary delete
// @Description update single data
// @Tags  meta Manager
// @Param	id		path 	string	true		"The key for staticblock"
// @Accept json
// @Produce json
// @Router /v1/images/delete/:id [delete]
func Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if id <= 0 || err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "id must be int type", err))
		return
	}
	err = models.DeleteImageMeta(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	util.Log.Warnf("The  ImageMeta (Id:%d) had been delete ", id)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", id))
}

// @Summary QueryMyHistory
// @Description Query My History
// @Tags  meta Manager
// @Accept json
// @Produce json
// @Router /v1/images/queryHistory/mine [get]
func QueryMyHistory(c *gin.Context) {
	//...... emplty . wait for query param
	var UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	if UserId <= 0 {
		c.JSON(http.StatusForbidden, util.ExportData(util.CodeStatusClientError, " forbidden ", nil))
		return
	}

	result, err := models.GetMyImageMetaHistory(UserId, 0, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
}
