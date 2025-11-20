import requests
import json
from rich.console import Console
from rich.traceback import install
install(show_locals=True)
console = Console()

url = "http://141.11.187.215:8080/process"
headers = {
    "Content-Type": "application/json"
}

# Your previos Tree
Tree = """
None
"""

# Your Prompt
prompt = """
A vhdl code for adder subtractor multiplier divider and a testbench for it
"""
data = {
    "Tree": Tree,
    "Prompt": prompt,
    "OperationType": "Edit"
}

response = requests.post(url, headers=headers, data=json.dumps(data))

try:
    result = response.json()  # Parse JSON response
    print("==Tree file==")
    tree = result.get("Tree", "No 'Tree' section found") 
    console.print(result,style="bold red")

    # print("==Model Questions==") # It doesnt know that asked you thease questions !he's memory less !
    # print(result.get("Question", "No 'Question' section found"))
except json.JSONDecodeError as e:
    print("Response is not valid JSON:")
    print(e)
    console.print(response.text,style="bold red")

