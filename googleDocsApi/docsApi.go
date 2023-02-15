package googledocsapi

import (
	"context"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
	"scrap/constant"
)


// ------------------------- Синхронизация таблицы в docs с базой знаний -------------------------
func Updatetable(srv *docs.Service, heading constant.Heading, scrapTable []constant.TableResponseCodes) error {
	tab, err := getTable(srv)
	if err != nil {
		return err
	}

	// если таблицы нет в docs, созадние новой
	if tab == nil {
		return CreatTable(srv, heading, scrapTable)
	}

	requests := []*docs.Request{}

	var nowText string
	// Если размерности таблицы совпадают то проверка и обновление измененных данных
	if len(tab.TableRows)-1 == len(scrapTable) && len(tab.TableRows[0].TableCells) == 2{
		for i := len(scrapTable) - 1; i >= 0; i-- {
			nowText = tab.TableRows[i+1].TableCells[1].Content[0].Paragraph.Elements[0].TextRun.Content
			nowText = nowText[:len(nowText)-1]

			if nowText != scrapTable[i].Desctiption {
				
				requests = append(requests, creatDeleteRequest(tab, i+1, 1))
				requests = append(requests, creatInsertRequest(tab, scrapTable[i].Desctiption, i+1, 1))
			}

			nowText = tab.TableRows[i+1].TableCells[0].Content[0].Paragraph.Elements[0].TextRun.Content
			nowText = nowText[:len(nowText)-1]
			if nowText != scrapTable[i].Code {
				requests = append(requests, creatDeleteRequest(tab, i+1, 0))
				requests = append(requests, creatInsertRequest(tab, scrapTable[i].Code, i+1, 0))
			}
		}
	} else {
		// Если размености не совпадают, тто удаление таблицы и создание новой с новыми даными и размером
		for i := 0; i < len(tab.TableRows); i++ {
			requests = append(requests, &docs.Request{
				DeleteTableRow: &docs.DeleteTableRowRequest{
					TableCellLocation: &docs.TableCellLocation{
						TableStartLocation: &docs.Location{ 
							Index: 2,
						},
						RowIndex: 0,
						ColumnIndex: 0,
					},
				},
			})
		}
		err := butchUpdate(srv, requests)
		if err != nil {
			return err
		}
		return CreatTable(srv, heading, scrapTable)
	} 


	// Проверка заголовка таблицы
	nowText = tab.TableRows[0].TableCells[1].Content[0].Paragraph.Elements[0].TextRun.Content
	nowText = nowText[:len(nowText)-1]
	if nowText != heading.Desctiption {
		requests = append(requests, creatDeleteRequest(tab, 0, 1))
		requests = append(requests, creatInsertRequest(tab, heading.Desctiption, 0, 1))
	}

	nowText = tab.TableRows[0].TableCells[0].Content[0].Paragraph.Elements[0].TextRun.Content
	nowText = nowText[:len(nowText)-1]
	if nowText != heading.Code {
		requests = append(requests, creatDeleteRequest(tab, 0, 0))
		requests = append(requests, creatInsertRequest(tab, heading.Code, 0, 0))
	}

	if len(requests) > 0 {
		err := butchUpdate(srv, requests)
		if err != nil {
			return err
		}
	}

	return nil
}

// ------------------------- Создание таблицы в docs -------------------------
func CreatTable(srv *docs.Service, heading constant.Heading, scrapTable []constant.TableResponseCodes) error {
	table := docs.InsertTableRequest{}
	table.Location = &docs.Location{Index: 1}
	table.Columns = 2
	table.Rows = int64(len(scrapTable)) + 1

	requests := []*docs.Request{}
	requests = append(requests, &docs.Request{InsertTable: &table})

	err := butchUpdate(srv, requests)
	if err != nil {
		return err
	}
	requests = []*docs.Request{}

	tab, err := getTable(srv)
	if err != nil {
		return err
	}

	for i := len(scrapTable) - 1; i >= 0; i-- {
		requests = append(requests, creatInsertRequest(tab, scrapTable[i].Desctiption, i+1, 1))
		requests = append(requests, creatInsertRequest(tab, scrapTable[i].Code, i+1, 0))
	}

	requests = append(requests, creatInsertRequest(tab, heading.Desctiption, 0, 1))
	requests = append(requests, creatInsertRequest(tab, heading.Code, 0, 0))

	err = butchUpdate(srv, requests)
	if err != nil {
		return err
	}
	requests = []*docs.Request{}

	tab, err = getTable(srv)
	if err != nil {
		return err
	}

	requests = append(requests, &docs.Request{
		UpdateTextStyle: &docs.UpdateTextStyleRequest{
			TextStyle: &docs.TextStyle{Bold: true},
			Range: &docs.Range{
				StartIndex: tab.TableRows[0].TableCells[0].Content[0].StartIndex,
				EndIndex:   tab.TableRows[0].TableCells[0].Content[0].EndIndex,
			},
			Fields: "bold",
		},
	})

	requests = append(requests, &docs.Request{
		UpdateTextStyle: &docs.UpdateTextStyleRequest{
			TextStyle: &docs.TextStyle{Bold: true},
			Range: &docs.Range{
				StartIndex: tab.TableRows[0].TableCells[1].Content[0].StartIndex,
				EndIndex: tab.TableRows[0].TableCells[1].Content[0].EndIndex,
			},
			Fields: "bold",
		},
	})

	requests = append(requests, &docs.Request{
		UpdateTableCellStyle: &docs.UpdateTableCellStyleRequest{
			TableCellStyle: &docs.TableCellStyle{
				BackgroundColor: &docs.OptionalColor{
					Color: &docs.Color{
						RgbColor: &docs.RgbColor{
							Red: 0.85,
							Green: 0.85,
							Blue: 0.85,
						},
					},
				},
			},
			Fields: "backgroundColor",
			TableRange: &docs.TableRange{
				RowSpan: 1,
				ColumnSpan: 2,
				TableCellLocation: &docs.TableCellLocation{
					ColumnIndex: 0,
					RowIndex: 0,
					TableStartLocation: &docs.Location{ 
						Index: 2,
					},
				},
			},
		},
	})

	requests = append(requests, &docs.Request{
		UpdateTableColumnProperties: &docs.UpdateTableColumnPropertiesRequest{
			ColumnIndices: []int64{0},
			TableStartLocation: &docs.Location{
				Index: 2,
			},
			TableColumnProperties: &docs.TableColumnProperties{
				WidthType: "FIXED_WIDTH",
				Width: &docs.Dimension{
					Magnitude: 120,
					Unit: "PT",
				},
			},
			Fields: "*",
		},
	})

	err = butchUpdate(srv, requests)
	if err != nil {
		return err
	}

	return nil
}

// ------------------------- Подключение к google docs -------------------------
func GetService(nameFile string) (*docs.Service, error) {
	ctx := context.Background()

	credBytes, _ := os.ReadFile(nameFile)

	config, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/documents")
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx)

	srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}