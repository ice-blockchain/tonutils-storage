package db

import (
	"bytes"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/xssnick/tonutils-storage/storage"
)

func (s *Storage) GetFileIndex(bagId []byte, id uint32) (*storage.FileIndex, error) {
	if len(bagId) != 32 {
		panic("invalid bag id len, should be 32")
	}

	k := make([]byte, 3+32+4)
	copy(k, "fi:")
	copy(k[3:3+32], bagId)
	binary.LittleEndian.PutUint32(k[3+32:], id)

	res, err := s.db.Get(k, nil)
	if err != nil {
		return nil, err
	}

	return &storage.FileIndex{
		BlockFrom:       binary.LittleEndian.Uint32(res),
		BlockTo:         binary.LittleEndian.Uint32(res[4:]),
		BlockFromOffset: binary.LittleEndian.Uint32(res[8:]),
		BlockToOffset:   binary.LittleEndian.Uint32(res[12:]),
		Name:            string(res[16:]),
	}, nil
}

func (s *Storage) SetFileIndex(bagId []byte, id uint32, fi *storage.FileIndex) error {
	if len(bagId) != 32 {
		panic("invalid bag id len, should be 32")
	}

	k := make([]byte, 3+4+32)
	copy(k, "fi:")
	copy(k[3:3+32], bagId)
	binary.LittleEndian.PutUint32(k[3+32:], id)

	name := []byte(fi.Name)
	v := make([]byte, 16+len(name))
	binary.LittleEndian.PutUint32(v, fi.BlockFrom)
	binary.LittleEndian.PutUint32(v[4:], fi.BlockTo)
	binary.LittleEndian.PutUint32(v[8:], fi.BlockFromOffset)
	binary.LittleEndian.PutUint32(v[12:], fi.BlockToOffset)
	copy(v[16:], name)

	return s.db.Put(k, v, nil)
}

func (s *Storage) GetPiece(bagId []byte, id uint32) (*storage.PieceInfo, error) {
	if len(bagId) != 32 {
		panic("invalid bag id len, should be 32")
	}

	k := make([]byte, 3+32+4)
	copy(k, "pc:")
	copy(k[3:3+32], bagId)
	binary.LittleEndian.PutUint32(k[3+32:], id)

	res, err := s.db.Get(k, nil)
	if err != nil {
		return nil, err
	}

	return &storage.PieceInfo{
		StartFileIndex: binary.LittleEndian.Uint32(res),
		Proof:          res[4:],
	}, nil
}

func (s *Storage) RemovePiece(bagId []byte, id uint32) error {
	if len(bagId) != 32 {
		panic("invalid bag id len, should be 32")
	}

	k := make([]byte, 3+32+4)
	copy(k, "pc:")
	copy(k[3:3+32], bagId)
	binary.LittleEndian.PutUint32(k[3+32:], id)

	return s.db.Delete(k, nil)
}

func (s *Storage) SetPiece(bagId []byte, id uint32, p *storage.PieceInfo) error {
	if len(bagId) != 32 {
		panic("invalid bag id len, should be 32")
	}

	k := make([]byte, 3+32+4)
	copy(k, "pc:")
	copy(k[3:3+32], bagId)
	binary.LittleEndian.PutUint32(k[3+32:], id)

	v := make([]byte, 4+len(p.Proof))
	binary.LittleEndian.PutUint32(v, p.StartFileIndex)
	copy(v[4:], p.Proof)

	return s.db.Put(k, v, nil)
}

func (s *Storage) PiecesMask(bagId []byte, num uint32) []byte {
	if len(bagId) != 32 {
		panic("invalid bag id len, should be 32")
	}

	k := make([]byte, 3+32)
	copy(k, "pc:")
	copy(k[3:3+32], bagId)

	p := num / 8
	if num%8 != 0 {
		p++
	}

	mask := make([]byte, p)
	iter := s.db.NewIterator(&util.Range{Start: k}, nil)
	for iter.Next() {
		if !bytes.HasPrefix(iter.Key(), k) {
			break
		}
		id := binary.LittleEndian.Uint32(iter.Key()[len(k):])
		mask[id/8] |= 1 << (id % 8)
	}
	return mask
}