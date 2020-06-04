package cache

import (
	"io"
	"time"
	"unsafe"
)

var (
	_ = unsafe.Sizeof(0)
	_ = io.ReadFull
	_ = time.Now()
)

type Response struct {
	Status int16
	Hname  []string
	Hvals  [][]string
	Rdata  []byte
}

func (d *Response) Size() (s uint64) {

	{
		l := uint64(len(d.Hname))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Hname {

			{
				l := uint64(len(d.Hname[k0]))

				{

					t := l
					for t >= 0x80 {
						t >>= 7
						s++
					}
					s++

				}
				s += l
			}

		}

	}
	{
		l := uint64(len(d.Hvals))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Hvals {

			{
				l := uint64(len(d.Hvals[k0]))

				{

					t := l
					for t >= 0x80 {
						t >>= 7
						s++
					}
					s++

				}

				for k1 := range d.Hvals[k0] {

					{
						l := uint64(len(d.Hvals[k0][k1]))

						{

							t := l
							for t >= 0x80 {
								t >>= 7
								s++
							}
							s++

						}
						s += l
					}

				}

			}

		}

	}
	{
		l := uint64(len(d.Rdata))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 2
	return
}
func (d *Response) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.Status >> 0)

		buf[1+0] = byte(d.Status >> 8)

	}
	{
		l := uint64(len(d.Hname))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+2] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+2] = byte(t)
			i++

		}
		for k0 := range d.Hname {

			{
				l := uint64(len(d.Hname[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+2] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+2] = byte(t)
					i++

				}
				copy(buf[i+2:], d.Hname[k0])
				i += l
			}

		}
	}
	{
		l := uint64(len(d.Hvals))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+2] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+2] = byte(t)
			i++

		}
		for k0 := range d.Hvals {

			{
				l := uint64(len(d.Hvals[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+2] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+2] = byte(t)
					i++

				}
				for k1 := range d.Hvals[k0] {

					{
						l := uint64(len(d.Hvals[k0][k1]))

						{

							t := uint64(l)

							for t >= 0x80 {
								buf[i+2] = byte(t) | 0x80
								t >>= 7
								i++
							}
							buf[i+2] = byte(t)
							i++

						}
						copy(buf[i+2:], d.Hvals[k0][k1])
						i += l
					}

				}
			}

		}
	}
	{
		l := uint64(len(d.Rdata))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+2] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+2] = byte(t)
			i++

		}
		copy(buf[i+2:], d.Rdata)
		i += l
	}
	return buf[:i+2], nil
}

func (d *Response) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Status = 0 | (int16(buf[i+0+0]) << 0) | (int16(buf[i+1+0]) << 8)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+2] & 0x7F)
			for buf[i+2]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+2]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Hname)) >= l {
			d.Hname = d.Hname[:l]
		} else {
			d.Hname = make([]string, l)
		}
		for k0 := range d.Hname {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+2] & 0x7F)
					for buf[i+2]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+2]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				d.Hname[k0] = string(buf[i+2 : i+2+l])
				i += l
			}

		}
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+2] & 0x7F)
			for buf[i+2]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+2]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Hvals)) >= l {
			d.Hvals = d.Hvals[:l]
		} else {
			d.Hvals = make([][]string, l)
		}
		for k0 := range d.Hvals {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+2] & 0x7F)
					for buf[i+2]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+2]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				if uint64(cap(d.Hvals[k0])) >= l {
					d.Hvals[k0] = d.Hvals[k0][:l]
				} else {
					d.Hvals[k0] = make([]string, l)
				}
				for k1 := range d.Hvals[k0] {

					{
						l := uint64(0)

						{

							bs := uint8(7)
							t := uint64(buf[i+2] & 0x7F)
							for buf[i+2]&0x80 == 0x80 {
								i++
								t |= uint64(buf[i+2]&0x7F) << bs
								bs += 7
							}
							i++

							l = t

						}
						d.Hvals[k0][k1] = string(buf[i+2 : i+2+l])
						i += l
					}

				}
			}

		}
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+2] & 0x7F)
			for buf[i+2]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+2]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Rdata)) >= l {
			d.Rdata = d.Rdata[:l]
		} else {
			d.Rdata = make([]byte, l)
		}
		copy(d.Rdata, buf[i+2:])
		i += l
	}
	return i + 2, nil
}
