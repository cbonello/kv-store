import socket

# Good enough for my test project. Check https://stackoverflow.com/questions/319279/how-to-validate-ip-address-in-python
# if you want to go further.
def isValidIP(ip):
    ad = ip.split(":")
    if len(ad) > 2:
        return False
    try:
        socket.inet_aton(ad[0])
    except socket.error:
        return False
    if len(ad) > 1:
        # ":" not followed by a port number?
        if len(ad[1])>0:
            try:
                _ = int(ad[1])
                return True
            except:
                return False
    # IP address without a port number.
    return False
