package archiver

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const tempDir = "data/temp_recordings"

type TTSChunkMetadata struct {
	FilePath string `json:"file_path"`
	StartMS  int64  `json:"start_ms"`
}

type ArchiveS2CJob struct {
	Data      []byte
	StartTime time.Duration
}

// c2s: client to server, s2c: server to client
type Archiver struct {
	sessionID        string
	baseTrackFile    *os.File
	baseTrackPath    string
	ttsChunkMetadata []TTSChunkMetadata
	ttsChunkCounter  atomic.Uint64
}

func NewArchiver(sessionID string) (*Archiver, error) {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("NewArchiver(): failed to create temp directory: %v", err)
	}
	baseTrackPath := filepath.Join(tempDir, fmt.Sprintf("%s_c2s.webm", sessionID))
	baseFile, err := os.Create(baseTrackPath)
	if err != nil {
		return nil, err
	}

	log.Printf("NewArchiver(): Created temp files session %s, %s", sessionID, baseTrackPath)
	return &Archiver{
		sessionID:        sessionID,
		baseTrackFile:    baseFile,
		baseTrackPath:    baseTrackPath,
		ttsChunkMetadata: make([]TTSChunkMetadata, 0), // 빈 슬라이스로 초기화
	}, nil
}

// C->S chunk를 임시 파일에 기록
func (a *Archiver) WriteC2S(chunk []byte) {
	if _, err := a.baseTrackFile.Write(chunk); err != nil {
		log.Printf("Archiver.WriteC2S(): failed to write c2s chunk: %v", err)
	}
}

// S->C chunk를 임시 파일에 기록
func (a *Archiver) WriteS2C(job ArchiveS2CJob) {
	count := a.ttsChunkCounter.Add(1)
	chunkFileName := fmt.Sprintf("%s_tts_chunk_%d.raw", a.sessionID, count)
	chunkFilePath := filepath.Join(tempDir, chunkFileName)

	if err := os.WriteFile(chunkFilePath, job.Data, 0644); err != nil {
		log.Printf("Archiver.WriteS2C(): failed to save S2C chunk %s: %v", chunkFileName, err)
		return
	}

	metadata := TTSChunkMetadata{
		FilePath: chunkFilePath,
		StartMS:  job.StartTime.Milliseconds(),
	}
	a.ttsChunkMetadata = append(a.ttsChunkMetadata, metadata)
	log.Printf("Archiver.WriteS2C(): Saved S2C chunk %s (Start Time: %dms) ", chunkFileName, metadata.StartMS)
}

func (a *Archiver) CloseBaseTrack() {
	if a.baseTrackFile != nil {
		a.baseTrackFile.Close()
	}
}

func (a *Archiver) MergeAndSave(finalFilePath string) error {
	log.Printf("Archiver.MergeAndSave(): Merging files for session %s to %s", a.sessionID, finalFilePath)
	// 1. 기본 입력: C->S (base_track.webm)
	args := []string{
		"-y", // 덮어쓰기
		"-f", "webm", "-i", a.baseTrackPath,
	}

	// 2. S->C(TTS) 청크 파일들을 입력으로 추가
	for _, chunk := range a.ttsChunkMetadata {
		args = append(args,
			"-f", "s16le", "-ar", "16000", "-ac", "1", // TTS 포맷 (Linear16)
			"-i", chunk.FilePath,
		)
	}

	// 3. 필터(filter_complex) 문자열 동적 생성
	var filterBuilder strings.Builder

	// 3a. TTS 청크들에 딜레이(adelay)를 적용하여 [t1], [t2]... 스트림 생성
	// 예: [1:a]adelay=5800|5800[t1]; [2:a]adelay=15200|15200[t2]; ...
	ttsInputs := ""
	for i, chunk := range a.ttsChunkMetadata {
		streamIndex := i + 1 // [0:a]는 C->S이므로, TTS 청크는 [1:a]부터 시작
		delay := chunk.StartMS
		filterBuilder.WriteString(fmt.Sprintf(
			"[%d:a]adelay=%d|%d,afade=type=in:duration=1[t%d]; ",
			streamIndex, delay, delay, i,
		))
		ttsInputs += fmt.Sprintf("[t%d]", i)
	}

	// 3b. C->S(기본) 트랙과 모든 딜레이된 TTS 트랙을 믹싱
	// 예: [0:a][t0][t1][t2]amix=inputs=4[final_mix]
	filterBuilder.WriteString(fmt.Sprintf("[0:a]%samix=inputs=%d[final_mix]", ttsInputs, len(a.ttsChunkMetadata)+1))

	args = append(args, "-filter_complex", filterBuilder.String())

	// 4. 최종 출력 설정
	args = append(args,
		"-map", "[final_mix]",
		"-c:a", "libmp3lame",
		"-q:a", "4",
		finalFilePath,
	)

	// 5. FFmpeg 명령어 실행
	cmd := exec.Command("ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Archiver: FFmpeg timestamp merge failed for %s. Error: %v. Output: %s", a.sessionID, err, string(output))
		// (임시 파일 삭제 로직은 에러와 상관없이 실행)
	} else {
		log.Printf("Archiver: Merge successful for %s", a.sessionID)
	}

	// 6. 임시 파일 삭제
	os.Remove(a.baseTrackPath)
	for _, chunk := range a.ttsChunkMetadata {
		os.Remove(chunk.FilePath)
	}
	log.Printf("Archiver: Temp files deleted for %s", a.sessionID)

	return err
}
