package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"NotesServer/gates/storage"
	"NotesServer/models/dto"
	"NotesServer/pkg"
	"net/http"
)

type HttpServer struct {
	srv http.Server
	st  storage.Storage
}

func NewHttpServer(addr string, st storage.Storage) *HttpServer {
	hs := &HttpServer{
		srv: http.Server{Addr: addr},
		st:  st,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/create", hs.recordCreateHandler)
	mux.HandleFunc("/get", hs.recordsGetHandler)
	mux.HandleFunc("/update", hs.recordUpdateHandler)
	mux.HandleFunc("/delete", hs.recordDeleteByPhone)
	mux.HandleFunc("/get-all", hs.recordGetAll)
	hs.srv.Handler = mux

	return hs
}

func (hs *HttpServer) Start() error {
	eW := pkg.NewEWrapper("(hs *HttpServer) Start()")
	defer eW.Close()

	if err := hs.srv.ListenAndServe(); err != nil {
		return eW.WrapError(err, "hs.srv.ListenAndServe()")
	}
	return nil
}

func (hs *HttpServer) recordCreateHandler(w http.ResponseWriter, req *http.Request) {
	setHeaders(w)

	eW, err := pkg.NewEWrapperWithFile("(hs *HttpServer) recordCreateHandler()")
	if err != nil {
		log.Println("(hs *HttpServer) recordCreateHandler: NewEWrapperWithFile()", err)
	}
	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	record := dto.NewNote()
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(req.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(req)")
		return
	}

	if record.Name == "" || record.LastName == "" || record.Note == "" {
		err = errors.New("required data is missing")
		resp.Wrap("Required data is missing", nil, err.Error())
		eW.LogError(err, "json.Unmarshal")
		return
	}
	record.ID = hs.st.NextIndex()
	idx, err := hs.st.Add(record)

	if err != nil {
		resp.Wrap("Error in saving record", nil, err.Error())
		eW.LogError(err, "hs.db.RecordSave(record)")
		return
	}

	idxMap := map[string]interface{}{
		"id": idx,
	}
	idxJson, err := json.Marshal(idxMap)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Marshal(idx)")
		return
	}

	resp.Wrap("Successfully added", idxJson, "")
}

func (hs *HttpServer) recordsGetHandler(w http.ResponseWriter, req *http.Request) {
	setHeaders(w)

	eW, err := pkg.NewEWrapperWithFile("(hs *HttpServer) recordsGetHandler()")
	if err != nil {
		log.Println("(hs *HttpServer) recordCreateHandler: NewEWrapperWithFile()", err)
	}
	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	record := dto.NewNote()
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(req.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(req)")
		return
	}

	if record.ID == -1 {
		err = errors.New("no ID provided")
		resp.Wrap("No ID provided", nil, err.Error())
		eW.LogError(err, "No ID provided")
		return
	}

	records, status := hs.st.GetByIndex(record.ID)
	if !status {
		resp.Wrap("Error in finding records", nil, errors.New("no records found").Error())
		eW.LogError(err, "hs.db.RecordsGet(record)")
		return
	}

	recordsJSON, err := json.Marshal(records)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Marshal(records)")
		return
	}

	resp.Wrap("Success", recordsJSON, "")
}

func (hs *HttpServer) recordUpdateHandler(w http.ResponseWriter, req *http.Request) {
	setHeaders(w)

	eW, err := pkg.NewEWrapperWithFile("(hs *HttpServer) recordUpdateHandler()")
	if err != nil {
		log.Println("(hs *HttpServer) recordCreateHandler: NewEWrapperWithFile()", err)
	}

	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	record := dto.NewNote()
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(req.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(byteReq, &record)")
		return
	}

	if (record.Name == "" && record.LastName == "" && record.Note == "") || record.ID == 0 {
		err = errors.New("required data is missing")
		resp.Wrap("Required data is missing", nil, err.Error())
		eW.LogError(err, "json.Unmarshal")
		return
	}

	hs.st.RemoveByIndex(record.ID)
	err = hs.st.AddToIndex(record, record.ID)
	if err != nil {
		resp.Wrap("Error in updating record", nil, err.Error())
		eW.LogError(err, "hs.db.RecordUpdate(record)")
		return
	}
	resp.Wrap("Success", nil, "")
}

func (hs *HttpServer) recordDeleteByPhone(w http.ResponseWriter, req *http.Request) {
	setHeaders(w)

	eW, err := pkg.NewEWrapperWithFile("(hs *HttpServer) recordDeleteByPhone()")
	if err != nil {
		log.Println("(hs *HttpServer) recordCreateHandler: NewEWrapperWithFile()", err)
	}

	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	record := dto.NewNote()
	byteReq, err := io.ReadAll(req.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(r.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(byteReq, &record)")
		return
	}

	if record.ID == -1 {
		err = errors.New("id is missing")
		resp.Wrap("ID is missing", nil, err.Error())
		eW.LogError(err, "json.Unmarshal")
		return
	}

	hs.st.RemoveByIndex(record.ID)
	resp.Wrap("Success", nil, "")
}

func (hs *HttpServer) recordGetAll(w http.ResponseWriter, req *http.Request) {
	setHeaders(w)

	eW, err := pkg.NewEWrapperWithFile("(hs *HttpServer) recordGetAll()")
	if err != nil {
		log.Println("(hs *HttpServer) recordCreateHandler: NewEWrapperWithFile()", err)
	}

	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	records, status := hs.st.GetAll()
	if !status {
		resp.Wrap("Error in finding records", nil, errors.New("no records found").Error())
		eW.LogError(err, "hs.db.RecordsGetAll()")
		return
	}

	recordsJSON, err := json.Marshal(records)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Marshal(records)")
		return
	}

	resp.Wrap("Success", recordsJSON, "")
}

func responseReturn(w http.ResponseWriter, eW *pkg.EWrapper, resp *dto.Response) {
	errEncode := json.NewEncoder(w).Encode(resp)
	if errEncode != nil {
		eW.LogError(errEncode, "json.NewEncoder(w).Encode(resp)")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	eW.Close()
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
