package vision

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type WordArray struct {
	Text          string
	WordPosition  int
	LineNumber    int
	RegionsNumber int
}

type ImgText struct {
	Words    []WordArray
	Language string
	Text     string
}

type ocrWords struct {
	BoundingBox string `json:"boundingBox"`
	Text        string `json:"text"`
}

type ocrLines struct {
	BoundingBox string     `json:"boundingBox"`
	Words       []ocrWords `json:"words"`
}

type ocrRegions struct {
	BoundingBox string     `json:"boundingBox"`
	Lines       []ocrLines `json:"lines"`
}

type ocrResponse struct {
	TextAngle   float64      `json:"textAngle"`
	Orientation string       `json:"orientation"`
	Language    string       `json:"language"`
	Regions     []ocrRegions `json:"regions"`
}

func getOcrImgText(jsonBody []byte, imgtext *ImgText) error {

	var ocrResp ocrResponse
	err := json.Unmarshal(jsonBody, &ocrResp)
	if err != nil {
		log.Info(err)
		return err
	}

	*imgtext = ImgText{}
	imgtext.Language = ocrResp.Language

	for rIt, rOcr := range ocrResp.Regions {
		for lIt, lOcr := range rOcr.Lines {
			for wIt, wOcr := range lOcr.Words {
				imgtext.Words = append(imgtext.Words,
					WordArray{Text: wOcr.Text,
						WordPosition:  wIt,
						LineNumber:    lIt,
						RegionsNumber: rIt})
				imgtext.Text += wOcr.Text + " "
			}
			imgtext.Text += "\n"
		}
	}

	return nil
}

type ocrRespError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func getVisionErrorMsg(jsonBody []byte) (string, error) {
	var errResp ocrRespError
	err := json.Unmarshal(jsonBody, &errResp)
	if err != nil {
		log.Info(err)
		return "", err
	}

	errMsg := "Vision Error: Error code: " + errResp.Code + " messege: " + errResp.Message
	return errMsg, nil

}

type Vision struct {
	ServerUrl string
	ApiKey    string
	Client    *http.Client
}

func CreateVision(serverUrl string, apiKey string) Vision {
	return Vision{serverUrl, apiKey, &http.Client{}}
}

func (imgt *Vision) GetTextFromImg(imgPath string, lang string) (*ImgText, error) {

	urlV := url.Values{}
	urlV.Set("subscription-key", imgt.ApiKey)

	imgT := &ImgText{}

	var requestMethod string

	urlV.Set("lang", lang)
	urlV.Set("detectOrientation", "true")
	requestMethod = "ocr"

	fullServerPath := imgt.ServerUrl + "/" + requestMethod

	urlPath, err := url.ParseRequestURI(fullServerPath)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	var dataBuff *bytes.Buffer
	var contentTypeRequest string

	if strings.HasPrefix(imgPath, "file://") {
		var imgData []byte
		imgPath = strings.TrimPrefix(imgPath, "file://")
		imgData, err = ioutil.ReadFile(imgPath)
		if err != nil {
			log.Info(err)
			return nil, err
		}
		dataBuff = bytes.NewBuffer(imgData)
		contentTypeRequest = "application/octet-stream"
	} else {
		var imgUrl *url.URL
		imgUrl, err = url.ParseRequestURI(imgPath)
		if err != nil {
			log.Info(err)
			return nil, err
		}
		dataBuff = bytes.NewBufferString("{\"url\":" + "\"" + imgUrl.String() + "\"}")
		contentTypeRequest = "application/json"
	}

	urlPath.RawQuery = urlV.Encode()

	visionResponse, err := imgt.Client.Post(urlPath.String(), contentTypeRequest, dataBuff)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	defer visionResponse.Body.Close()

	responseBody, err := ioutil.ReadAll(visionResponse.Body)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	if visionResponse.StatusCode != 200 {
		errMsg, err := getVisionErrorMsg(responseBody)
		if err != nil {

			log.Info(err)
			return nil, err
		}
		err = errors.New(errMsg)
		log.Info(err)
		return nil, err
	}

	err = getOcrImgText(responseBody, imgT)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	return imgT, nil
}
