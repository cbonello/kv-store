const checkIp = require('check-ip')

const isValidIP = ip => {
  const ad = ip.split(':') // checkIp does not handle ports...
  if (ad.length > 2) {
    return false
  }
  const host = ad[0] === 'localhost' ? '127.0.0.1' : ad[0]
  if (!checkIp(host).isValid) {
    return false
  }
  if (ad.length > 1) {
    // ':' not followed by a port number?
    if (ad[1].length > 0) {
      const port = parseInt(ad[1])
      if (isNaN(port)) {
        return false
      }
      return true
    }
  }
  // IP address without a port number.
  return false
}

module.exports = isValidIP
