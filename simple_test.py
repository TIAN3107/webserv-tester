import urllib.request, subprocess, time, os, signal, socket

# name = input("Type your name ")
# if (name == "Timo") :
# 	print ("Hallo Timo")
# else :
# 	print ("Who tf are you??.. anyways")
# 	name = input("Obviously ")

#SET THIS to your server adress
serverAdress = "http://127.0.0.1:8080"

print ("Hello World")
# print ("Spamming single requests")
cmd = ["valgrind", "--leak-check=full", "--track-fds=yes", "./webserv"]
server = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
# while socket.create_connection()
# while server.poll() == TimeoutError :
# 	time.sleep(0.5)
print ("Did it work?")
time.sleep(3)
print ("Done sleeping")
i = 0
while i < 1000 :
	contents = urllib.request.urlopen(serverAdress).read()
	print (contents)
	print(i, "\n")
	print(f"{contents} {i} \n")
	i+=1
	time.sleep(0.1)
# time.sleep(20)
server.send_signal(signal.SIGINT)
output, _ = server.communicate(timeout=30)
print(output)
print("exit code:", server.returncode)