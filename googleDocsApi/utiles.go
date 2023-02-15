package googledocsapi
import (
	"google.golang.org/api/docs/v1"
	"scrap/constant"
)
// внесение измененией 
func butchUpdate(srv *docs.Service, requests []*docs.Request) error {
	batch := docs.BatchUpdateDocumentRequest{}
	batch.Requests = requests

	_, err := srv.Documents.BatchUpdate(constant.DocId, &batch).Do()
	return err
}

// получение таблицы из docs
func getTable(srv *docs.Service) (*docs.Table, error) {
	doc, err := srv.Documents.Get(constant.DocId).Do()
	if err != nil {
		return nil, err
	}

	content := doc.Body.Content

	var table *docs.Table
	for i := 0; i < len(content) && table == nil; i++ {
		if content[i].Table != nil {
			table = content[i].Table
		}
	}
	return table, nil
}

// создание запроса на вствку в таблицу
func creatInsertRequest(table *docs.Table, text string, posI, posJ int) *docs.Request {
	return &docs.Request{
		InsertText: &docs.InsertTextRequest{
			Text: text,
			Location: &docs.Location{
				Index: table.TableRows[posI].TableCells[posJ].Content[0].StartIndex,
			},
		},
	}
}

// создание запроса на удаление из таблицы
func creatDeleteRequest(table *docs.Table, posI, posJ int) *docs.Request {
	start :=  table.TableRows[posI].TableCells[posJ].Content[0].StartIndex
	end := table.TableRows[posI].TableCells[posJ].Content[0].EndIndex - 1
	if start - end != 0 {
		return &docs.Request{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: start,
					EndIndex:   end,
				},
			},
		}
	}
	return nil
}