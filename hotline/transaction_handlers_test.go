package hotline

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func TestHandleSetChatSubject(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    []Transaction
		wantErr bool
	}{
		{
			name: "sends chat subject to private chat members",
			args: args{
				cc: &ClientConn{
					UserName: []byte{0x00, 0x01},
					Server: &Server{
						PrivateChats: map[uint32]*PrivateChat{
							uint32(1): {
								Subject: "unset",
								ClientConn: map[uint16]*ClientConn{
									uint16(1): {
										Account: &Account{
											Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
										},
										ID: &[]byte{0, 1},
									},
									uint16(2): {
										Account: &Account{
											Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
										},
										ID: &[]byte{0, 2},
									},
								},
							},
						},
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 1},
							},
							uint16(2): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 2},
							},
						},
					},
				},
				t: &Transaction{
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0, 0, 0, 1},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
						NewField(fieldChatSubject, []byte("Test Subject")),
					},
				},
			},
			want: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x77},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
						NewField(fieldChatSubject, []byte("Test Subject")),
					},
				},
				{
					clientID:  &[]byte{0, 2},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x77},
					ID:        []byte{0xf0, 0xc5, 0x34, 0x1e}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
						NewField(fieldChatSubject, []byte("Test Subject")),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		rand.Seed(1) // reset seed between tests to make transaction IDs predictable

		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleSetChatSubject(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleSetChatSubject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("HandleSetChatSubject() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleLeaveChat(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    []Transaction
		wantErr bool
	}{
		{
			name: "returns expected transactions",
			args: args{
				cc: &ClientConn{
					ID: &[]byte{0, 2},
					Server: &Server{
						PrivateChats: map[uint32]*PrivateChat{
							uint32(1): {
								ClientConn: map[uint16]*ClientConn{
									uint16(1): {
										Account: &Account{
											Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
										},
										ID: &[]byte{0, 1},
									},
									uint16(2): {
										Account: &Account{
											Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
										},
										ID: &[]byte{0, 2},
									},
								},
							},
						},
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 1},
							},
							uint16(2): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 2},
							},
						},
					},
				},
				t: NewTransaction(tranDeleteUser, nil, NewField(fieldChatID, []byte{0, 0, 0, 1})),
			},
			want: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x76},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
						NewField(fieldUserID, []byte{0, 2}),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		rand.Seed(1)
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleLeaveChat(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleLeaveChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("HandleLeaveChat() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleGetUserNameList(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    []Transaction
		wantErr bool
	}{
		{
			name: "replies with userlist transaction",
			args: args{
				cc: &ClientConn{

					ID: &[]byte{1, 1},
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								ID:       &[]byte{0, 1},
								Icon:     &[]byte{0, 2},
								Flags:    &[]byte{0, 3},
								UserName: []byte{0, 4},
								Agreed:   true,
							},
							uint16(2): {
								ID:       &[]byte{0, 2},
								Icon:     &[]byte{0, 2},
								Flags:    &[]byte{0, 3},
								UserName: []byte{0, 4},
								Agreed:   true,
							},
							uint16(3): {
								ID:       &[]byte{0, 3},
								Icon:     &[]byte{0, 2},
								Flags:    &[]byte{0, 3},
								UserName: []byte{0, 4},
								Agreed:   false,
							},
						},
					},
				},
				t: &Transaction{
					ID:   []byte{0, 0, 0, 1},
					Type: []byte{0, 1},
				},
			},
			want: []Transaction{
				{
					clientID:  &[]byte{1, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 1},
					ID:        []byte{0, 0, 0, 1},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(
							fieldUsernameWithInfo,
							[]byte{00, 01, 00, 02, 00, 03, 00, 02, 00, 04},
						),
						NewField(
							fieldUsernameWithInfo,
							[]byte{00, 02, 00, 02, 00, 03, 00, 02, 00, 04},
						),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleGetUserNameList(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleGetUserNameList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandleChatSend(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    []Transaction
		wantErr bool
	}{
		{
			name: "sends chat msg transaction to all clients",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessSendChat)
							access := bits[:]
							return &access
						}(),
					},
					UserName: []byte{0x00, 0x01},
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 1},
							},
							uint16(2): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 2},
							},
						},
					},
				},
				t: &Transaction{
					Fields: []Field{
						NewField(fieldData, []byte("hai")),
					},
				},
			},
			want: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldData, []byte{0x0d, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x00, 0x01, 0x3a, 0x20, 0x20, 0x68, 0x61, 0x69}),
					},
				},
				{
					clientID:  &[]byte{0, 2},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0xf0, 0xc5, 0x34, 0x1e}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldData, []byte{0x0d, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x00, 0x01, 0x3a, 0x20, 0x20, 0x68, 0x61, 0x69}),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranChatSend, &[]byte{0, 1},
					NewField(fieldData, []byte("hai")),
				),
			},
			want: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to participate in chat.")),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sends chat msg as emote if fieldChatOptions is set",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessSendChat)
							access := bits[:]
							return &access
						}(),
					},
					UserName: []byte("Testy McTest"),
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 1},
							},
							uint16(2): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 2},
							},
						},
					},
				},
				t: &Transaction{
					Fields: []Field{
						NewField(fieldData, []byte("performed action")),
						NewField(fieldChatOptions, []byte{0x00, 0x01}),
					},
				},
			},
			want: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldData, []byte("\r*** Testy McTest performed action")),
					},
				},
				{
					clientID:  &[]byte{0, 2},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0xf0, 0xc5, 0x34, 0x1e},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldData, []byte("\r*** Testy McTest performed action")),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "only sends chat msg to clients with accessReadChat permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessSendChat)
							access := bits[:]
							return &access
						}(),
					},
					UserName: []byte{0x00, 0x01},
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 1},
							},
							uint16(2): {
								Account: &Account{
									Access: &[]byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
								ID: &[]byte{0, 2},
							},
						},
					},
				},
				t: &Transaction{
					Fields: []Field{
						NewField(fieldData, []byte("hai")),
					},
				},
			},
			want: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldData, []byte{0x0d, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x00, 0x01, 0x3a, 0x20, 0x20, 0x68, 0x61, 0x69}),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "only sends private chat msg to members of private chat",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessSendChat)
							access := bits[:]
							return &access
						}(),
					},
					UserName: []byte{0x00, 0x01},
					Server: &Server{
						PrivateChats: map[uint32]*PrivateChat{
							uint32(1): {
								ClientConn: map[uint16]*ClientConn{
									uint16(1): {
										ID: &[]byte{0, 1},
									},
									uint16(2): {
										ID: &[]byte{0, 2},
									},
								},
							},
						},
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Access: &[]byte{255, 255, 255, 255, 255, 255, 255, 255},
								},
								ID: &[]byte{0, 1},
							},
							uint16(2): {
								Account: &Account{
									Access: &[]byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
								ID: &[]byte{0, 2},
							},
							uint16(3): {
								Account: &Account{
									Access: &[]byte{0, 0, 0, 0, 0, 0, 0, 0},
								},
								ID: &[]byte{0, 3},
							},
						},
					},
				},
				t: &Transaction{
					Fields: []Field{
						NewField(fieldData, []byte("hai")),
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
					},
				},
			},
			want: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
						NewField(fieldData, []byte{0x0d, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x00, 0x01, 0x3a, 0x20, 0x20, 0x68, 0x61, 0x69}),
					},
				},
				{
					clientID:  &[]byte{0, 2},
					Flags:     0x00,
					IsReply:   0x00,
					Type:      []byte{0, 0x6a},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldChatID, []byte{0, 0, 0, 1}),
						NewField(fieldData, []byte{0x0d, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x00, 0x01, 0x3a, 0x20, 0x20, 0x68, 0x61, 0x69}),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleChatSend(tt.args.cc, tt.args.t)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleChatSend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tranAssertEqual(t, tt.want, got)
		})
	}
}

func TestHandleGetFileInfo(t *testing.T) {
	rand.Seed(1) // reset seed between tests to make transaction IDs predictable

	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr bool
	}{
		{
			name: "returns expected fields when a valid file is requested",
			args: args{
				cc: &ClientConn{
					ID: &[]byte{0x00, 0x01},
					Server: &Server{
						Config: &Config{
							FileRoot: func() string {
								path, _ := os.Getwd()
								return path + "/test/config/Files"
							}(),
						},
					},
				},
				t: NewTransaction(
					tranGetFileInfo, nil,
					NewField(fieldFileName, []byte("testfile.txt")),
					NewField(fieldFilePath, []byte{0x00, 0x00}),
				),
			},
			wantRes: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xce},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldFileName, []byte("testfile.txt")),
						NewField(fieldFileTypeString, []byte("Text File")),
						NewField(fieldFileCreatorString, []byte("ttxt")),
						NewField(fieldFileComment, []byte{}),
						NewField(fieldFileType, []byte("TEXT")),
						NewField(fieldFileCreateDate, make([]byte, 8)),
						NewField(fieldFileModifyDate, make([]byte, 8)),
						NewField(fieldFileSize, []byte{0x0, 0x0, 0x0, 0x17}),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand.Seed(1) // reset seed between tests to make transaction IDs predictable

			gotRes, err := HandleGetFileInfo(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleGetFileInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Clear the file timestamp fields to work around problems running the tests in multiple timezones
			// TODO: revisit how to test this by mocking the stat calls
			gotRes[0].Fields[5].Data = make([]byte, 8)
			gotRes[0].Fields[6].Data = make([]byte, 8)
			if !assert.Equal(t, tt.wantRes, gotRes) {
				t.Errorf("HandleGetFileInfo() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestHandleNewFolder(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr bool
	}{
		{
			name: "without required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					accessCreateFolder,
					&[]byte{0, 0},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to create folders.")),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "when path is nested",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessCreateFolder)
							access := bits[:]
							return &access
						}(),
					},
					ID: &[]byte{0, 1},
					Server: &Server{
						Config: &Config{
							FileRoot: "/Files/",
						},
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							mfs.On("Mkdir", "/Files/aaa/testFolder", fs.FileMode(0777)).Return(nil)
							mfs.On("Stat", "/Files/aaa/testFolder").Return(nil, os.ErrNotExist)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranNewFolder, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFolder")),
					NewField(fieldFilePath, []byte{
						0x00, 0x01,
						0x00, 0x00,
						0x03,
						0x61, 0x61, 0x61,
					}),
				),
			},
			wantRes: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xcd},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
				},
			},
			wantErr: false,
		},
		{
			name: "when path is not nested",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessCreateFolder)
							access := bits[:]
							return &access
						}(),
					},
					ID: &[]byte{0, 1},
					Server: &Server{
						Config: &Config{
							FileRoot: "/Files",
						},
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							mfs.On("Mkdir", "/Files/testFolder", fs.FileMode(0777)).Return(nil)
							mfs.On("Stat", "/Files/testFolder").Return(nil, os.ErrNotExist)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranNewFolder, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFolder")),
				),
			},
			wantRes: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xcd},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
				},
			},
			wantErr: false,
		},
		{
			name: "when UnmarshalBinary returns an err",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessCreateFolder)
							access := bits[:]
							return &access
						}(),
					},
					ID: &[]byte{0, 1},
					Server: &Server{
						Config: &Config{
							FileRoot: "/Files/",
						},
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							mfs.On("Mkdir", "/Files/aaa/testFolder", fs.FileMode(0777)).Return(nil)
							mfs.On("Stat", "/Files/aaa/testFolder").Return(nil, os.ErrNotExist)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranNewFolder, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFolder")),
					NewField(fieldFilePath, []byte{
						0x00,
					}),
				),
			},
			wantRes: []Transaction{},
			wantErr: true,
		},
		{
			name: "fieldFileName does not allow directory traversal",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessCreateFolder)
							access := bits[:]
							return &access
						}(),
					},
					ID: &[]byte{0, 1},
					Server: &Server{
						Config: &Config{
							FileRoot: "/Files/",
						},
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							mfs.On("Mkdir", "/Files/testFolder", fs.FileMode(0777)).Return(nil)
							mfs.On("Stat", "/Files/testFolder").Return(nil, os.ErrNotExist)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranNewFolder, &[]byte{0, 1},
					NewField(fieldFileName, []byte("../../testFolder")),
				),
			},
			wantRes: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xcd},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
				},
			}, wantErr: false,
		},
		{
			name: "fieldFilePath does not allow directory traversal",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessCreateFolder)
							access := bits[:]
							return &access
						}(),
					},
					ID: &[]byte{0, 1},
					Server: &Server{
						Config: &Config{
							FileRoot: "/Files/",
						},
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							mfs.On("Mkdir", "/Files/foo/testFolder", fs.FileMode(0777)).Return(nil)
							mfs.On("Stat", "/Files/foo/testFolder").Return(nil, os.ErrNotExist)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranNewFolder, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFolder")),
					NewField(fieldFilePath, []byte{
						0x00, 0x02,
						0x00, 0x00,
						0x03,
						0x2e, 0x2e, 0x2f,
						0x00, 0x00,
						0x03,
						0x66, 0x6f, 0x6f,
					}),
				),
			},
			wantRes: []Transaction{
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xcd},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42}, // Random ID from rand.Seed(1)
					ErrorCode: []byte{0, 0, 0, 0},
				},
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotRes, err := HandleNewFolder(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleNewFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tranAssertEqual(t, tt.wantRes, gotRes) {
				t.Errorf("HandleNewFolder() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestHandleUploadFile(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr bool
	}{
		{
			name: "when request is valid and user has Upload Anywhere permission",
			args: args{
				cc: &ClientConn{
					Server: &Server{
						FileTransfers: map[uint32]*FileTransfer{},
					},
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessUploadFile)
							bits.Set(accessUploadAnywhere)
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranUploadFile, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFile")),
					NewField(fieldFilePath, []byte{
						0x00, 0x01,
						0x00, 0x00,
						0x03,
						0x2e, 0x2e, 0x2f,
					}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xcb},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldRefNum, []byte{0x52, 0xfd, 0xfc, 0x07}), // rand.Seed(1)
					},
				},
			},
			wantErr: false,
		},
		{
			name: "when user does not have required access",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						FileTransfers: map[uint32]*FileTransfer{},
					},
				},
				t: NewTransaction(
					tranUploadFile, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFile")),
					NewField(fieldFilePath, []byte{
						0x00, 0x01,
						0x00, 0x00,
						0x03,
						0x2e, 0x2e, 0x2f,
					}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to upload files.")), // rand.Seed(1)
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand.Seed(1)
			gotRes, err := HandleUploadFile(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleUploadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)

		})
	}
}

func TestHandleMakeAlias(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr bool
	}{
		{
			name: "with valid input and required permissions",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessMakeAlias)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Config: &Config{
							FileRoot: func() string {
								path, _ := os.Getwd()
								return path + "/test/config/Files"
							}(),
						},
						Logger: NewTestLogger(),
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							path, _ := os.Getwd()
							mfs.On(
								"Symlink",
								path+"/test/config/Files/foo/testFile",
								path+"/test/config/Files/bar/testFile",
							).Return(nil)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranMakeFileAlias, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFile")),
					NewField(fieldFilePath, EncodeFilePath(strings.Join([]string{"foo"}, "/"))),
					NewField(fieldFileNewPath, EncodeFilePath(strings.Join([]string{"bar"}, "/"))),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0xd1},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields:    []Field(nil),
				},
			},
			wantErr: false,
		},
		{
			name: "when symlink returns an error",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessMakeAlias)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Config: &Config{
							FileRoot: func() string {
								path, _ := os.Getwd()
								return path + "/test/config/Files"
							}(),
						},
						Logger: NewTestLogger(),
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							path, _ := os.Getwd()
							mfs.On(
								"Symlink",
								path+"/test/config/Files/foo/testFile",
								path+"/test/config/Files/bar/testFile",
							).Return(errors.New("ohno"))
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranMakeFileAlias, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFile")),
					NewField(fieldFilePath, EncodeFilePath(strings.Join([]string{"foo"}, "/"))),
					NewField(fieldFileNewPath, EncodeFilePath(strings.Join([]string{"bar"}, "/"))),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("Error creating alias")),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Config: &Config{
							FileRoot: func() string {
								path, _ := os.Getwd()
								return path + "/test/config/Files"
							}(),
						},
					},
				},
				t: NewTransaction(
					tranMakeFileAlias, &[]byte{0, 1},
					NewField(fieldFileName, []byte("testFile")),
					NewField(fieldFilePath, []byte{
						0x00, 0x01,
						0x00, 0x00,
						0x03,
						0x2e, 0x2e, 0x2e,
					}),
					NewField(fieldFileNewPath, []byte{
						0x00, 0x01,
						0x00, 0x00,
						0x03,
						0x2e, 0x2e, 0x2e,
					}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to make aliases.")),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleMakeAlias(tt.args.cc, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleMakeAlias(%v, %v)", tt.args.cc, tt.args.t)
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleGetUser(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when account is valid",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessOpenUser)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{
							"guest": {
								Login:    "guest",
								Name:     "Guest",
								Password: "password",
								Access:   &[]byte{1},
							},
						},
					},
				},
				t: NewTransaction(
					tranGetUser, &[]byte{0, 1},
					NewField(fieldUserLogin, []byte("guest")),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0x01, 0x60},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldUserName, []byte("Guest")),
						NewField(fieldUserLogin, negateString([]byte("guest"))),
						NewField(fieldUserPassword, []byte("password")),
						NewField(fieldUserAccess, []byte{1}),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranGetUser, &[]byte{0, 1},
					NewField(fieldUserLogin, []byte("nonExistentUser")),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to view accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when account does not exist",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessOpenUser)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranGetUser, &[]byte{0, 1},
					NewField(fieldUserLogin, []byte("nonExistentUser")),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("Account does not exist.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleGetUser(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleGetUser(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleDeleteUser(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when user exists",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessDeleteUser)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{
							"testuser": {
								Login:    "testuser",
								Name:     "Testy McTest",
								Password: "password",
								Access:   &[]byte{1},
							},
						},
						FS: func() *MockFileStore {
							mfs := &MockFileStore{}
							mfs.On("Remove", "Users/testuser.yaml").Return(nil)
							return mfs
						}(),
					},
				},
				t: NewTransaction(
					tranDeleteUser, &[]byte{0, 1},
					NewField(fieldUserLogin, negateString([]byte("testuser"))),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0x1, 0x5f},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields:    []Field(nil),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranDeleteUser, &[]byte{0, 1},
					NewField(fieldUserLogin, negateString([]byte("testuser"))),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to delete accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleDeleteUser(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleDeleteUser(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleGetMsgs(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "returns news data",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessNewsReadArt)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						FlatNews: []byte("TEST"),
					},
				},
				t: NewTransaction(
					tranGetMsgs, &[]byte{0, 1},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x65},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldData, []byte("TEST")),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranGetMsgs, &[]byte{0, 1},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to read news.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleGetMsgs(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleGetMsgs(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleNewUser(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranNewUser, &[]byte{0, 1},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to create new accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleNewUser(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleNewUser(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleListUsers(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					tranNewUser, &[]byte{0, 1},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to view accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleListUsers(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleListUsers(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleDownloadFile(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when user does not have required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{},
				},
				t: NewTransaction(tranDownloadFile, &[]byte{0, 1}),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to download files.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "with a valid file",
			args: args{
				cc: &ClientConn{
					Transfers: make(map[int][]*FileTransfer),
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessDownloadFile)
							access := bits[:]
							return &access
						}(),
					},
					Server: &Server{
						FileTransfers: make(map[uint32]*FileTransfer),
						Config: &Config{
							FileRoot: func() string { path, _ := os.Getwd(); return path + "/test/config/Files" }(),
						},
						Accounts: map[string]*Account{},
					},
				},
				t: NewTransaction(
					accessDownloadFile,
					&[]byte{0, 1},
					NewField(fieldFileName, []byte("testfile.txt")),
					NewField(fieldFilePath, []byte{0x0, 0x00}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x2},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields: []Field{
						NewField(fieldRefNum, []byte{0x52, 0xfd, 0xfc, 0x07}),
						NewField(fieldWaitingCount, []byte{0x00, 0x00}),
						NewField(fieldTransferSize, []byte{0x00, 0x00, 0x00, 0xa5}),
						NewField(fieldFileSize, []byte{0x00, 0x00, 0x00, 0x17}),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset the rand seed so that the random fieldRefNum will be deterministic
			rand.Seed(1)

			gotRes, err := HandleDownloadFile(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleDownloadFile(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleUpdateUser(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when action is create user without required permission",
			args: args{
				cc: &ClientConn{
					Server: &Server{
						Logger: NewTestLogger(),
					},
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranUpdateUser,
					&[]byte{0, 0},
					NewField(fieldData, []byte{
						0x00, 0x04, // field count

						0x00, 0x69, // fieldUserLogin = 105
						0x00, 0x03,
						0x9d, 0x9d, 0x9d,

						0x00, 0x6a, // fieldUserPassword = 106
						0x00, 0x03,
						0x9c, 0x9c, 0x9c,

						0x00, 0x66, // fieldUserName = 102
						0x00, 0x03,
						0x61, 0x61, 0x61,

						0x00, 0x6e, // fieldUserAccess = 110
						0x00, 0x08,
						0x60, 0x70, 0x0c, 0x20, 0x03, 0x80, 0x00, 0x00,
					}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to create new accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when action is modify user without required permission",
			args: args{
				cc: &ClientConn{
					Server: &Server{
						Logger: NewTestLogger(),
						Accounts: map[string]*Account{
							"bbb": {},
						},
					},
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranUpdateUser,
					&[]byte{0, 0},
					NewField(fieldData, []byte{
						0x00, 0x04, // field count

						0x00, 0x69, // fieldUserLogin = 105
						0x00, 0x03,
						0x9d, 0x9d, 0x9d,

						0x00, 0x6a, // fieldUserPassword = 106
						0x00, 0x03,
						0x9c, 0x9c, 0x9c,

						0x00, 0x66, // fieldUserName = 102
						0x00, 0x03,
						0x61, 0x61, 0x61,

						0x00, 0x6e, // fieldUserAccess = 110
						0x00, 0x08,
						0x60, 0x70, 0x0c, 0x20, 0x03, 0x80, 0x00, 0x00,
					}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to modify accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when action is delete user without required permission",
			args: args{
				cc: &ClientConn{
					Server: &Server{
						Logger: NewTestLogger(),
						Accounts: map[string]*Account{
							"bbb": {},
						},
					},
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranUpdateUser,
					&[]byte{0, 0},
					NewField(fieldData, []byte{
						0x00, 0x01,
						0x00, 0x65,
						0x00, 0x03,
						0x88, 0x9e, 0x8b,
					}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to delete accounts.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleUpdateUser(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleUpdateUser(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleDelNewsArt(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "without required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranDelNewsArt,
					&[]byte{0, 0},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to delete news articles.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleDelNewsArt(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleDelNewsArt(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}
			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleDisconnectUser(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "without required permission",
			args: args{
				cc: &ClientConn{
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranDelNewsArt,
					&[]byte{0, 0},
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("You are not allowed to disconnect users.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when target user has 'cannot be disconnected' priv",
			args: args{
				cc: &ClientConn{
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(1): {
								Account: &Account{
									Login: "unnamed",
									Access: func() *[]byte {
										var bits accessBitmap
										bits.Set(accessCannotBeDiscon)
										access := bits[:]
										return &access
									}(),
								},
							},
						},
					},
					Account: &Account{
						Access: func() *[]byte {
							var bits accessBitmap
							bits.Set(accessDisconUser)
							access := bits[:]
							return &access
						}(),
					},
				},
				t: NewTransaction(
					tranDelNewsArt,
					&[]byte{0, 0},
					NewField(fieldUserID, []byte{0, 1}),
				),
			},
			wantRes: []Transaction{
				{
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0, 0x00},
					ID:        []byte{0x9a, 0xcb, 0x04, 0x42},
					ErrorCode: []byte{0, 0, 0, 1},
					Fields: []Field{
						NewField(fieldError, []byte("unnamed is not allowed to be disconnected.")),
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleDisconnectUser(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleDisconnectUser(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}
			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}

func TestHandleSendInstantMsg(t *testing.T) {
	type args struct {
		cc *ClientConn
		t  *Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Transaction
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "when client 1 sends a message to client 2",
			args: args{
				cc: &ClientConn{
					ID:       &[]byte{0, 1},
					UserName: []byte("User1"),
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(2): {
								AutoReply: []byte(nil),
							},
						},
					},
				},
				t: NewTransaction(
					tranSendInstantMsg,
					&[]byte{0, 1},
					NewField(fieldData, []byte("hai")),
					NewField(fieldUserID, []byte{0, 2}),
				),
			},
			wantRes: []Transaction{
				*NewTransaction(
					tranServerMsg,
					&[]byte{0, 2},
					NewField(fieldData, []byte("hai")),
					NewField(fieldUserName, []byte("User1")),
					NewField(fieldUserID, []byte{0, 1}),
					NewField(fieldOptions, []byte{0, 1}),
				),
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0x0, 0x6c},
					ID:        []byte{0, 0, 0, 0},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields:    []Field(nil),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "when client 2 has autoreply enabled",
			args: args{
				cc: &ClientConn{
					ID:       &[]byte{0, 1},
					UserName: []byte("User1"),
					Server: &Server{
						Clients: map[uint16]*ClientConn{
							uint16(2): {
								ID:        &[]byte{0, 2},
								UserName:  []byte("User2"),
								AutoReply: []byte("autohai"),
							},
						},
					},
				},
				t: NewTransaction(
					tranSendInstantMsg,
					&[]byte{0, 1},
					NewField(fieldData, []byte("hai")),
					NewField(fieldUserID, []byte{0, 2}),
				),
			},
			wantRes: []Transaction{
				*NewTransaction(
					tranServerMsg,
					&[]byte{0, 2},
					NewField(fieldData, []byte("hai")),
					NewField(fieldUserName, []byte("User1")),
					NewField(fieldUserID, []byte{0, 1}),
					NewField(fieldOptions, []byte{0, 1}),
				),
				*NewTransaction(
					tranServerMsg,
					&[]byte{0, 1},
					NewField(fieldData, []byte("autohai")),
					NewField(fieldUserName, []byte("User2")),
					NewField(fieldUserID, []byte{0, 2}),
					NewField(fieldOptions, []byte{0, 1}),
				),
				{
					clientID:  &[]byte{0, 1},
					Flags:     0x00,
					IsReply:   0x01,
					Type:      []byte{0x0, 0x6c},
					ID:        []byte{0, 0, 0, 0},
					ErrorCode: []byte{0, 0, 0, 0},
					Fields:    []Field(nil),
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := HandleSendInstantMsg(tt.args.cc, tt.args.t)
			if !tt.wantErr(t, err, fmt.Sprintf("HandleSendInstantMsg(%v, %v)", tt.args.cc, tt.args.t)) {
				return
			}

			tranAssertEqual(t, tt.wantRes, gotRes)
		})
	}
}
