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

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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
	insertData.Packages = imageInputData.Packages
	insertData.Version = imageInputData.Version
	insertData.BuildType = imageInputData.BuildType
	if len(insertData.Version) == 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "verison not allowed empty ", nil))
		return
	}
	//check package validate
	validate := false
	for _, pkgs := range util.GetConfig().BuildParam.Packages {
		if pkgs == insertData.Packages {
			validate = true
			break
		}
	}
	if !validate {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "packages not supported  ", util.GetConfig().BuildParam.Packages))
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

	var temp []byte
	temp, err = json.Marshal(imageInputData.CustomPkg)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	insertData.CustomPkg = string(temp)
	//----------------------send data to k8s to build----
	controllerID := uuid.NewV4().String()
	var jobName = fmt.Sprintf(`omni-image-%s`, controllerID)
	var imageName = fmt.Sprintf(`openEuler-%s.iso`, controllerID)
	omniImager := `omni-imager --package-list /etc/omni-imager/` + insertData.Packages + `.json --config-file /etc/omni-imager/conf.yaml --build-type ` + insertData.BuildType + ` --output-file ` + imageName
	omniCurl := `curl -vvv -Ffile=@/opt/omni-workspace/` + imageName + ` -Fproject=` + insertData.Version + `  -FfileType=image 'https://repo.test.osinfra.cn/data/upload?token=316462d0c029ba707ad2'`
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata": map[string]interface{}{
				"name":      jobName,
				"namespace": util.GetConfig().K8sConfig.Namespace,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"job-name": jobName,
					},
				},
				"ttlSecondsAfterFinished": 1800,
				"backoffLimit":            1,
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"job-name": jobName,
						},
					},

					"spec": map[string]interface{}{
						"restartPolicy": "Never",
						"containers": []map[string]interface{}{
							{
								"name":    "image-builder",
								"image":   util.GetConfig().K8sConfig.Image,
								"command": []string{"/bin/sh", "-c", omniImager, omniCurl},
							},
						},
					},
				},
			},
		},
	}

	client, err := dynamic.NewForConfig(models.GetK8sConfig())
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "Create job Error", err))
		return
	}
	deploymentRes := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	deploy, err := client.Resource(deploymentRes).Namespace(util.GetConfig().K8sConfig.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "Create job Error", err))
		return
	}

	insertData.JobName = deploy.GetName()
	insertData.CreateTime = deploy.GetCreationTimestamp().Time
	jobDBID, err := models.AddImageMeta(&insertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, jobDBID, deploy.GetName(), util.GetConfig().WSConfig))
}

// @Summary QueryJobStatus
// @Description QueryJobStatus for given job name
// @Tags  meta Manager
// @Param	name		path 	string	true		"The name for job"
// @Param	id		query 	string	false		"The id for job in database. "
// @Accept json
// @Produce json
// @Router /v1/images/queryJobStatus/{name} [get]
func QueryJobStatus(c *gin.Context) {
	jobname := c.Param("name")
	if jobname == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job name must be fill:", nil))
		return
	}
	jobidStr, _ := c.GetQuery("id")
	// if given jobid . update job status in database
	jobid, _ := strconv.Atoi(jobidStr)
	jobAPI := models.GetClientSet().BatchV1()
	job, err := jobAPI.Jobs(util.GetConfig().K8sConfig.Namespace).Get(context.TODO(), jobname, metav1.GetOptions{})
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
	if job.Status.Succeeded > *completions {
		result["status"] = models.JOB_STATUS_SUCCEED
		result["completionTime"] = job.Status.CompletionTime
		result["url"] = fmt.Sprintf(util.GetConfig().BuildParam.DownloadIsoUrl, jobname)
		if jobid > 0 {
			var updateJob models.ImageMeta
			updateJob.JobName = jobname
			updateJob.Id = jobid
			updateJob.Status = models.JOB_STATUS_SUCCEED
			models.UpdateJobStatus(&updateJob)
		}
	} else if job.Status.Failed > *backoffLimit {
		result["status"] = models.JOB_STATUS_FAILED
		result["completionTime"] = job.Status.CompletionTime
		if jobid > 0 {
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
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", util.GetConfig().BuildParam, util.GetConfig().DefaultPkgList))
}

// @Summary GetCustomePkgList param
// @Description get default package name list. this list load from https://raw.githubusercontent.com/omnibuildplatform/omni-imager/main/etc/openEuler-minimal.json
// @Tags  meta Manager
// @Accept json
// @Produce json
// @Router /v1/images/param/getCustomePkgList/ [get]
func GetCustomePkgList(c *gin.Context) {
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", nil))
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
