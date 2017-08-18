package vision

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type WordArray struct {
	Text          string
	WordPosition  uint64
	LineNumber    uint64
	RegionsNumber uint64
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

func getOcrImgText(jsonBody []byte, imgtext *ImgText) (err error) {

	var ocrResp ocrResponse
	err = json.Unmarshal(jsonBody, &ocrResp)
	if err != nil {
		return
	}

	imgtext = ImgText{}
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

	return
}

type hwWords struct {
	BoundingBox []int  `json:"boundingBox"`
	Text        string `json:"text"`
}

type hwLines struct {
	BoundingBox []int     `json:"boundingBox"`
	Text        string    `json:"text"`
	Words       []hwWords `json:"words"`
}

type hwRecognitionResult struct {
	Lines []hwLines `json:"lines"`
}

type hwResponse struct {
	Status            string              `json:"status"`
	Succeeded         bool                `json:"succeeded"`
	Failed            bool                `json:"failed"`
	Finished          bool                `json:"finished"`
	RecognitionResult hwRecognitionResult `json:"recognitionResult"`
}

func getHwImgText(jsonBody []byte, imgtext *ImgText) (err error) {

	var hwResp hwResponse
	err = json.Unmarshal(jsonBody, &hwResp)
	if err != nil {
		return
	}

	imgtext = ImgText{}
	imgtext.Language = "en"

	for lIt, lHw := range hwResp.RecognitionResult.Lines {
		for wIt, wHw := range lHw.Words {
			imgtext.Words = append(imgtext.Words,
				WordArray{Text: wHw.Text,
					WordPosition:  wIt,
					LineNumber:    lIt,
					RegionsNumber: 0})
			imgtext.Text += wHw.Text + " "
		}
		imgtext.Text += "\n"
	}

	return
}

const (
	FullPathType uint8 = iota
	UrlPathType
	OcrImgType
	HwImgType
)

type Vision struct {
	ServerUrl string
	ApiKey    string
	Client    *http.Client
	Text      ImgText
}

func CreateVisoin(serverUrl string, apiKey string) Vision {
	return Vision{serverUrl, apiKey, &http.Client{}, ImgText{}}
}

func (imgt *Vision) GetTextFromImg(imgPath string, pathType uint8, recImgType uint8i, lang string) (err error) {

	urlV := url.Values{}
	urlV.Set("lang", lang)
	urlV.Set("subscription-key", imgt.ApiKey)

	var requestMethod string
	var responseParseFunction func([]byte, *ImgText) error

	if recImgType == OcrImgType {
		urlV.Set("detectOrientation", "true")
		requestMethod = "ocr"
		responseParseFunction = getOcrImgText
	} else if recImgType == HwImgType {
		urlV.Set("handwriting", "true")
		requestMethod = "recognizeText"
		responseParseFunction = getHwImgText
	} else {
		err = errors.New("invalide recImgType")
		return
	}

	fullServerPath = imgt.ServerUrl + requestMethod

	urlPath, err := url.ParseRequestURI(fullServerPath)
	if err != nil {
		return
	}

	var dataBuff *Buffer
	var contentTypeRequest string

	if pathType == FullPathType {
		imgData, err := ioutil.ReadFile(imgPath)
		if err != nil {
			return
		}
		dataBuff = bytes.NewBuffer(imgData)
		contentTypeRequest = "application/octet-stream"
	} else if pathType == UrlPathType {
		imgUrl, err = url.ParseRequestURI(imgPath)
		if err != nil {
			return
		}
		dataBuff = bytes.NewBufferString("{\"url\":" + "\"" + imgUrl.String() + "\"}")
		contentTypeRequest = "application/json"
	} else {
		err = errors.New("invalide pathType")
		return
	}

	urlPath.RawQuery = urlV.Encode()

	visionResponse, err := imgt.Client.Post(urlPath.String(), contentTypeRequest, dataBuff)
	if err != nil {
		return
	}
	defer visionResponse.Body.Close()

	if visionResponse.StatusCode != 200 {
		err = errors.New("Vision response failed satus code: " + visionResponse.Status)
		return
	}

	responseBody, err := ioutil.ReadAll(visionResponse.Body)
	if err != nil {
		return
	}

	err = responseParseFunction(responseBody, &imgt.Text)
	if err != nil {
		return
	}

	return
}
