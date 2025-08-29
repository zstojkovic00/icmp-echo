package icmp

/*
ICMP Echo Request/Reply Message Format (RFC 792)

    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |     Type      |     Code      |          Checksum             |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |           Identifier          |        Sequence Number        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |     Data ...
   +-+-+-+-+-

Type:
  8 for Echo Request
  0 for Echo Reply

Code:
  0

Checksum:
  16-bit one's complement of the one's complement sum of the ICMP message
  starting with the ICMP Type. For computing the checksum, this field is zero.

Identifier:
  An identifier to aid in matching echo requests and replies, may be zero.

Sequence Number:
  A sequence number to aid in matching echo requests and replies, may be zero.

Data:
  The data received in the echo message must be returned in the echo reply.
*/

// ICMP message types according to RFC 792
const (
	EchoRequest = 8 // Echo Request message
	EchoReply   = 0 // Echo Reply message
)

// Message represents an ICMP Echo Request or Reply message
type Message struct {
	Type           uint8  // Message type (8 = request, 0 = reply)
	Code           uint8  // Message code (always 0 for echo)
	Checksum       uint16 // Internet checksum of ICMP message
	Identifier     uint16
	SequenceNumber uint16
	Data           []byte // Optional data payload
}

func NewEchoRequest(identifier uint16, data []byte) *Message {
	return &Message{
		Type:           EchoRequest,
		Code:           0,
		Checksum:       0,
		Identifier:     identifier,
		SequenceNumber: 0,
		Data:           data,
	}
}

func (m *Message) Serialize() []byte {
	length := 8 + len(m.Data) // 8-byte header + data
	packet := make([]byte, length)

	// ICMP Header (8 bytes)
	packet[0] = m.Type
	packet[1] = m.Code
	packet[2] = 0
	packet[3] = 0
	packet[4] = 0
	packet[5] = 0
	packet[6] = 0
	packet[7] = 0

	// Copy data payload
	copy(packet[8:], m.Data)

	// Calculate and set checksum
	checksum := CalculateChecksum(packet)
	packet[2] = byte(checksum >> 8)
	packet[3] = byte(checksum & 0xFF)

	return packet
}

/*
	Packet: [8, 0, 0, 0, 0, 0, 0, 0]

	i=0: pair = 8*256 + 0 = 2048,    checksum = 2048
	i=2: pair = 0*256 + 0 = 0,       checksum = 2048
	i=4: pair = 0*256 + 0 = 0,       checksum = 2048
	i=6: pair = 0*256 + 0 = 0,       checksum = 2048
*/

func CalculateChecksum(packet []byte) uint16 {
	var checksum uint32

	for i := 0; i < len(packet); i += 2 {
		if i+1 < len(packet) {
			pair := uint32(packet[i])*256 + uint32(packet[i+1])
			checksum += pair
		} else {
			checksum += uint32(packet[i]) * 256
		}
	}

	return uint16(^checksum)
}
