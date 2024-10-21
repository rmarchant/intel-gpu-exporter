package intel_gpu_top

import (
	"bytes"
	"io"
)

var _ io.Reader = &JSONFixer{}

type JSONFixer struct {
	Reader    io.Reader
	buffer    bytes.Buffer
	skipFirst bool
}

func (j *JSONFixer) Read(p []byte) (int, error) {
	// if the buffer is not empty, drain it first
	if j.buffer.Len() > 0 {
		return j.buffer.Read(p)
	}

	tmp := make([]byte, len(p))
	n, err := j.Reader.Read(tmp)
	if err != nil {
		return n, err
	}
	if n > 0 {
		data := tmp[:n]

		// we will be prepending the start of every new record with a comma, so:
		// 	{ <stats> }
		//	{ <stats> }
		// becomes:
		// 	{ <stats> },
		//	{ <stats> }
		// we therefore need to skip the start of the first record

		if !j.skipFirst {
			// Find the start of the next record
			if index := bytes.Index(data, []byte("\n{")); index != -1 {
				j.buffer.Write(data[:index+1])
				data = data[index+1:]
				j.skipFirst = true
			}
		}

		// we're beyond the first record. prepend the start of the next record with a comma.
		replaced := bytes.ReplaceAll(data, []byte("\n{"), []byte(",\n{"))
		j.buffer.Write(replaced)
	}

	return j.buffer.Read(p)
}
