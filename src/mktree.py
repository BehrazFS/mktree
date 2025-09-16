#!/usr/bin/env python3
import os, sys

# Re-run using venv’s Python if not already inside it
project_path = os.path.dirname(__file__)[:-4]
venv_python = (project_path + "/.venv/bin/python3")
if sys.executable != venv_python:
    os.execv(venv_python, [venv_python] + sys.argv)

import argparse
from typing import List,Dict,Optional,Tuple
from dataclasses import dataclass, field
from rich.traceback import install
from rich import print as rprint
from rich.text import Text
from rich.console import Console
from rich.tree import Tree
install(show_locals=True)
console = Console()
verbosity_level = 1

def get_arguments() -> Dict[str,str]:
    """ Parse command-line arguments and return them as a dictionary."""
    parser = argparse.ArgumentParser(
        description="mktree – create directory/file structures from a .tree file or inline definition."
    )

    # Positional argument
    parser.add_argument(
        "tree_file",
        nargs="?",
        default=None,
        help="Path to the .tree file (optional if using --inline or --interactive)"
    )

    # Optional arguments
    parser.add_argument("-i", "--inline", type=str, help="Inline tree structure as a string.") # done
    parser.add_argument("-d", "--dry-run", action="store_true", help="Show structure without creating files/folders.") #done
    parser.add_argument("-f", "--force", action="store_true", help="Overwrite existing files or directories.") # done
    parser.add_argument("-c", "--clean", action="store_true", help="Remove all files/folders specified in the tree.") # not implemented in this version
    parser.add_argument("-V", "--verbose", action="store_true", help="Print detailed info about operations.") # done
    parser.add_argument("-q", "--quiet", action="store_true", help="Suppress all output except errors.") # done
    parser.add_argument("-t", "--template", type=str, help="Use a built-in template (e.g., python, latex, cpp).")
    parser.add_argument("-I", "--interactive", action="store_true", help="Launch interactive mode.") # not implemented in this version (coming soon)
    parser.add_argument("-o", "--output", type=str, default=".", help="Base directory to create the tree in (default: current directory).") # done
    parser.add_argument("-r", "--reverse", action="store_true", help="Generate a .tree file from an existing directory.")
    parser.add_argument("-v", "--version", action="version", version="mktree 1.0.0", help="Show program version.") # done
    parser.add_argument("-n", "--no-content", action="store_true",help="When reversing, do not include file contents in the .tree output.") # done

    args = parser.parse_args()
    return vars(args)

@dataclass
class Node:
    name: str
    type: str  # 'dir' or 'file'
    children: List['Node'] = field(default_factory=list)
    content: Optional[str] = None  # Only for files

def count_leading_spaces(line: str) -> int:
    """ Count leading spaces in a line."""
    return len(line) - len(line.lstrip(' '))

def parse_entry(line: str) -> Tuple[str, Optional[str]]:
    """
    Returns (name, content) tuple.
    If line contains ':', split name and content.
    """
    if ':|' in line:  # multiline content
        name, _ = line.split(':|', 1)
        return name.strip(), ':|'
    elif ':' in line:  # single-line content
        name, content = line.split(':', 1)
        return name.strip(), content.strip()
    else:
        return line.strip(), None

def detect_type(name: str, content: Optional[str]) -> str:
    """ 
    Detect if entry is a file or directory.
    If content is provided or name has an extension, it's a file.
    Otherwise, it's a directory.
    """
    if content is not None or '.' in name:
        return 'file'
    else:
        return 'dir'

def read_multiline_content(lines: List[str], start_index: int, base_indent: int) -> Tuple[str, int]:
    """
    Read lines that are part of a multiline content block.
    Stops when indentation is <= base_indent
    Returns (content_string, next_index)
    """
    content_lines: List[str] = []
    i: int = start_index
    while i < len(lines):
        line: str = lines[i]
        # stop if dedented
        if count_leading_spaces(line) // 2 <= base_indent:
            break
        # remove base_indent spaces and extra 2 spaces for content
        content_lines.append(line[base_indent * 2 + 2:])  
        i += 1
    return ''.join(content_lines), i

def parse_tree_file(file_path: str) -> Node:
    """ Parse a .tree file and return the root Node. """
    with open(file_path, 'r') as f:
        lines = f.readlines()

    root_node: Node = Node(name="$ROOT", type="dir")
    stack: List[Node] = [root_node]
    i: int = 0

    while i < len(lines):
        line: str = lines[i]
        if line.strip() == "" or line.strip().startswith("#"):
            i += 1
            continue

        indent: int = count_leading_spaces(line) // 2
        name, content_indicator = parse_entry(line)

        node_type: str = detect_type(name, content_indicator)
        node: Node = Node(name=name, type=node_type)

        if content_indicator == ':|':  # multiline content
            node.content, next_i = read_multiline_content(lines, i+1, indent)
            i = next_i
        elif content_indicator is not None:  # single-line content
            node.content = content_indicator
            i += 1
        else:
            i += 1

        # adjust stack to correct parent
        while indent < len(stack) - 1:
            stack.pop()

        parent: Node = stack[-1]
        parent.children.append(node)

        if node.type == "dir":
            stack.append(node)

    return root_node

def parse_tree_inline(tree: str) -> Node:
    """ Parse an inline tree string and return the root Node. """
    lines = tree.splitlines()
    root_node: Node = Node(name="$ROOT", type="dir")
    stack: List[Node] = [root_node]
    i: int = 0

    while i < len(lines):
        line: str = lines[i]
        if line.strip() == "" or line.strip().startswith("#"):
            i += 1
            continue

        indent: int = count_leading_spaces(line) // 2
        name, content_indicator = parse_entry(line)

        node_type: str = detect_type(name, content_indicator)
        node: Node = Node(name=name, type=node_type)

        if content_indicator == ':|':  # multiline content
            node.content, next_i = read_multiline_content(lines, i+1, indent)
            i = next_i
        elif content_indicator is not None:  # single-line content
            node.content = content_indicator
            i += 1
        else:
            i += 1

        # adjust stack to correct parent
        while indent < len(stack) - 1:
            stack.pop()

        parent: Node = stack[-1]
        parent.children.append(node)

        if node.type == "dir":
            stack.append(node)

    return root_node

# Deprecated: simple print function
def print_tree(node, prefix=""):
    """ Recursively print the tree structure to console.(Deprecated) """
    icon : str = "📁" if node.type == "dir" else "📄"
    print(f"{prefix}{icon} {node.name}")
    for child in node.children:
        print_tree(child, prefix + "  ")
        if child.content:
            content_lines = child.content.split('\n')
            for line in content_lines:
                print(f"{prefix}    📝 {line}")

def build_rich_tree(node: Node) -> Tree:
    """
    Recursively convert Node hierarchy into a Rich Tree.
    """
    label = f"📁[bold blue]{node.name}[/]" if node.type == "dir" else f"📄[green]{node.name}[/]"
    tree = Tree(label, guide_style="bold bright_blue") if node.type == "dir" else Tree(label, guide_style="green")
    
    for child in node.children:
        if child.type == "dir":
            tree.add(build_rich_tree(child))
        else:
            if child.content:
                # For files with content, add the content as child nodes
                content_label = "\n".join([f"[purple]{str(i).rjust(5,' ')}[/] [pink]|[/] [yellow]{line}[/]" for i, line in enumerate(child.content.splitlines(), start=1)])
                tree.add(f"📄[green]{child.name}[/]\n{content_label}")
            else:
                tree.add(f"📄[green]{child.name}[/]")
    return tree

def create_tree(node: Node, base_path: str, force:bool = False) -> None:
    """
    Recursively create directories and files on disk from Node tree.
    Supports --force, and --output.

    """
    full_path = os.path.join(base_path, node.name) if node.name != "$ROOT" else base_path
    if node.name != "$ROOT":
        if node.type == "dir":
            os.makedirs(full_path, exist_ok=True)
        else: 
            file_exists = os.path.exists(full_path)
            if file_exists and not force:
                log(f"[red]Error:[/] File '{full_path}' already exists. Use --force to overwrite.", v_level=1)
                raise FileExistsError(f"File '{full_path}' exists.")
            else:    
                os.makedirs(os.path.dirname(full_path), exist_ok=True)
                with open(full_path, "w", encoding="utf-8") as f:
                    if node.content:
                        f.write(node.content)
    for child in node.children:
        create_tree(child, full_path, force=force)

def is_binary_file(path: str, blocksize: int = 1024) -> bool:
    """ heuristic to detect binary files"""
    with open(path, "rb") as f:
        chunk = f.read(blocksize)
    # If it has null bytes, it's probably binary
    if b"\0" in chunk:
        return True
    # Heuristic: ratio of non-printable chars
    text_chars = bytearray({7, 8, 9, 10, 12, 13, 27} | set(range(0x20, 0x100)))
    nontext = chunk.translate(None, text_chars)
    return float(len(nontext)) / max(len(chunk), 1) > 0.30


def build_tree_from_directory(path: str, no_content: bool = False, size_limit: int = 1_000_000) -> Node:
    """
    Recursively walk a directory and return a Node tree.
    """
    name = os.path.basename(path.rstrip(os.sep))
    if not name:
        name = path  # root fallback

    node = Node(name=name, type="dir")

    try:
        for entry in sorted(os.listdir(path)):
            full_path = os.path.join(path, entry)
            if os.path.isdir(full_path):
                node.children.append(build_tree_from_directory(full_path, no_content, size_limit))
            else:
                if no_content:
                    content = None
                elif is_binary_file(full_path):
                    content = "<binary file>"
                elif os.path.getsize(full_path) < size_limit:
                    with open(full_path, "r", encoding="utf-8", errors="ignore") as f:
                        content = f.read()
                else:
                    content = None
                node.children.append(Node(name=entry, type="file", content=content))
    except PermissionError:
            log(f"[red]Warning:[/] Skipping {path} (permission denied)", v_level=1)
    return node

def node_to_tree_lines(node: Node, indent: int = 0) -> List[str]:
    """
    Convert a Node tree into .tree file format lines.
    """
    lines: List[str] = []
    if node.name != "$ROOT":  # skip synthetic root
        prefix = "  " * indent
        if node.type == "dir":
            lines.append(f"{prefix}{node.name}")
        else:
            if node.content and "\n" in node.content:
                lines.append(f"{prefix}{node.name}:|")
                for line in node.content.splitlines():
                    lines.append(f"{prefix}  {line}")
            elif node.content:
                lines.append(f"{prefix}{node.name}: {node.content}")
            else:
                lines.append(f"{prefix}{node.name}")

    for child in node.children:
        lines.extend(node_to_tree_lines(child, indent + (0 if node.name == "$ROOT" else 1)))

    return lines

def log(message: str, v_level: int = 1) -> None:
    if v_level > verbosity_level: return
    console.print(message)


if __name__ == "__main__":
    try:
        args: Dict[str, str] = get_arguments()
        if args.get("quiet", False):
            verbosity_level = 0
        if args.get("verbose", False):
            verbosity_level = 2
        if args.get("clean", False):
            raise NotImplementedError("--clean not implemented yet.")
        if args.get("interactive", False):
            raise NotImplementedError("--interactive not implemented yet.")
        
        if verbosity_level == 2 : rprint(args) # Debug print of arguments
        if args.get("reverse", False):
            if not args['tree_file']:
                raise ValueError("Please provide a directory path to reverse using the positional argument.")
            dir_path = args['tree_file']
            if not os.path.isdir(dir_path):
                raise NotADirectoryError(f"'{dir_path}' is not a valid directory.")
            tree = build_tree_from_directory(dir_path, no_content=args.get("no_content", False))
            tree_lines = node_to_tree_lines(tree)
            output_str = "\n".join(tree_lines)
            log(f"[bold green]Generated .tree structure from '{dir_path}':[/]\n", v_level=2)
            log(output_str, v_level=2)
            if not args.get("dry_run", False):
                output_file = os.path.join(args.get("output", "."), f"{os.path.basename(dir_path)}.tree")
                with open(output_file, "w", encoding="utf-8") as f:
                    f.write(output_str)
                log(f"\n[bold blue]Saved to:[/] {output_file}", v_level=2)
            sys.exit(0)
        if args['template']:
            template_path: str = project_path + "/templates"
            if not os.path.isdir(template_path):
                raise NotADirectoryError(f"Templates directory '{template_path}' not found.")
            template_file = os.path.join(template_path, f"{args['template'].lower()}.tree")
            if not os.path.isfile(template_file):
                raise FileNotFoundError(f"Template '{args['template']}' not found in templates directory.")
            tree = parse_tree_file(template_file)
        elif args['inline']:
            tree = parse_tree_inline(args['inline'])
        elif args['tree_file']:
            tree = parse_tree_file(args['tree_file'])
        else:
            raise NotImplementedError("Either --inline or tree_file must be provided.")
            # TODO: implement interactive mode
        
        # print_tree(tree)
        rich_tree = build_rich_tree(tree)
        console.print(rich_tree)
        if not args.get("dry_run", False):
            from rich.prompt import Confirm
            correct_tree = Confirm.ask("Is this tree structure correct?")
            if not correct_tree:
                raise SystemExit("Aborted by user.")
            create_tree(tree, base_path=args.get("output", "."), force=args.get("force", False))
    except NotImplementedError as e:
        log(f"[red]Error:[/] {e}", v_level=1)
        sys.exit(1)
    except SystemExit as e:
        log(f"[yellow]{e}[/]", v_level=1)
        sys.exit(0)
    except FileExistsError as e:
        sys.exit(1)
    except ValueError as e:
        log(f"[red]Error:[/] {e}",v_level=1)
        sys.exit(1)
    except NotADirectoryError as e:
        log(f"[red]Error:[/] {e}", v_level=1)
        sys.exit(1)
    except FileNotFoundError as e:
        log(f"[red]Error:[/] {e}", v_level=1)
        sys.exit(1)
    except PermissionError as e:
        log(f"[red]Error:[/] {e}", v_level=1)
        sys.exit(1)