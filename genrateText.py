import random
import string
import os
def generate_word(word_length = 5):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=word_length))
    
def generate_line(word_count= 12):
    words = [ generate_word() for _ in range(word_count)] 
    words.append("\n")
    return " ".join(words)

def generate_paragraph(line_count, fileName):
    with open(fileName, 'w') as file:
        for  _ in range(line_count):
            file.write(generate_line())    
    file.close()

if __name__ == "__main__":
    fileNames = [("small.txt", 100), ("medium.txt", 1000), ("large.txt", 100000), ("xl.txt", 1000000), ("xll.txt", 10000000), ("xlll.txt", 100000000)] 
    # fileNames = [("5.txt", 50000000), ("2.txt", 20000000)] 
    # fileNames = [("xll.txt", 10000000)] 

    for fileName, line_count in fileNames:
        dir = "text"
        os.makedirs(dir, exist_ok=True)
        fileName = os.path.join(dir, fileName)
        generate_paragraph(line_count, fileName)
