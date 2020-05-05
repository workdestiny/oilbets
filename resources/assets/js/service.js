function _base64ToArrayBuffer(base64) {
  var binaryString = window.atob(base64)
  var len = binaryString.length
  var bytes = new Uint8Array(len)
  for (var i = 0; i < len; i++) {
    bytes[i] = binaryString.charCodeAt(i)
  }
  return bytes.buffer
}

export default {
  _base64ToArrayBuffer
}