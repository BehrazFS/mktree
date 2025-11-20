import os
import json
import threading
import subprocess
from pathlib import Path
from kivy.uix.widget import Widget
import requests
from kivy.clock import mainthread
from kivy.core.window import Window
from kivymd.uix.screen import MDScreen
from kivymd.uix.screenmanager import MDScreenManager
from kivy.lang import Builder
from kivy.properties import StringProperty, BooleanProperty, ObjectProperty,NumericProperty
from kivymd.app import MDApp
from kivymd.uix.button import MDButton, MDButtonText
from kivymd.uix.dialog import MDDialog
from kivymd.uix.boxlayout import MDBoxLayout
from kivymd.uix.textfield import MDTextField
from kivymd.uix.label import MDLabel
from kivymd.uix.list import  MDList
from kivymd.uix.navigationrail import MDNavigationRailItem
from kivymd.uix.scrollview import MDScrollView
from kivymd.uix.navigationrail import MDNavigationRailItem
from kivy.metrics import dp
from kivymd.app import MDApp
from kivymd.uix.filemanager import MDFileManager
from kivymd.uix.snackbar import MDSnackbar, MDSnackbarText
from kivymd.uix.dialog import (
    MDDialog,
    MDDialogIcon,
    MDDialogHeadlineText,
    MDDialogSupportingText,
    MDDialogButtonContainer,
    MDDialogContentContainer,
)
from kivymd.uix.divider import MDDivider
from kivymd.uix.list import (
    MDListItem,
    MDListItemLeadingIcon,
    MDListItemSupportingText,
)

project_path = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
CONFIG_PATH = os.path.join(project_path, "src", "configs", ".mktree_gui_config.json")
LLM_API_URL_PLACEHOLDER = '<INSERT_SERVER_URL_HERE>'

# ---------- Window size ----------
def load_window_size():
    json_config_path = os.path.join(project_path, "src", "configs", "window_size.json")
    with open(json_config_path, 'r') as f:
        config = json.load(f)
        return config.get("width", 1100), config.get("height", 700)

Window.size = load_window_size()
# ---------- Error popup ----------
def show_error(msg, title='Error'):
    dialog = MDDialog(
        title=title,
        text=msg,
        size_hint=(0.5, 0.3),
        buttons=[MDButton(text="Close", on_release=lambda x: dialog.dismiss())]
    )
    dialog.open()

# ---------- Settings ----------
class Settings:
    def __init__(self, path=CONFIG_PATH):
        self.path = Path(path)
        self.data = {
            'llm_api_url': LLM_API_URL_PLACEHOLDER,
            'preview_colored': True,
            'editor_font_size': 14,
        }
        self.load()

    def load(self):
        if self.path.exists():
            try:
                with open(self.path, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    self.data.update(data)
            except Exception as e:
                show_error(f'Failed to load settings.\n{e}')

    def save(self):
        try:
            with open(self.path, 'w', encoding='utf-8') as f:
                json.dump(self.data, f, indent=2)
        except Exception as e:
            show_error(f'Failed to save settings.\n{e}')


settings = Settings()

# -----------------------------
class CommonNavigationRailItem(MDNavigationRailItem):
    text = StringProperty()
    icon = StringProperty()

# ---------- Screens ----------
class HomeScreen(MDScreen):
    pass

class TreeEditorScreen(MDScreen):
    editor = ObjectProperty(None)
    preview = ObjectProperty(None)
    def save_file(self,name):
        path = os.path.join(project_path, 'tree_files', f'{name!s}.tree')
        if not os.path.exists(os.path.dirname(path)):
            os.makedirs(os.path.dirname(path))
        try:
            with open(path, 'w', encoding='utf-8') as f:
                f.write(self.editor.text)
                MDSnackbar(
                    MDSnackbarText(
                        text="File saved successfully.",
                    ),
                    y=dp(24),
                    pos_hint={"center_x": 0.5},
                    size_hint_x=0.8,
                    duration=1.0 
                ).open()
        except Exception as e:
            MDSnackbar(
                MDSnackbarText(
                    text="Failed to save file.",
                ),
                y=dp(24),
                pos_hint={"center_x": 0.5},
                size_hint_x=0.8,
                duration=1.0 
            ).open()
    def show_alert_dialog(self):
        diag = MDDialog(
            # ----------------------------Icon-----------------------------
            MDDialogIcon(
                icon="content-save-check-outline",
            ),
            # -----------------------Headline text-------------------------
            MDDialogHeadlineText(
                text="Enter file name",
            ),
            # -----------------------Supporting text-----------------------
            MDDialogSupportingText(
                text="Choose a name for your tree file.",
            ),
            # -----------------------Custom content------------------------
            MDDialogContentContainer(
                MDDivider(),
                file_name_input := MDTextField(id="file_name_input", hint_text="File name", text="untitled"),
                MDDivider(),
                orientation="vertical",
            ),
            # ---------------------Button container------------------------
            MDDialogButtonContainer(
                Widget(),
                MDButton(
                    MDButtonText(text="Cancel"),
                    style="text",
                    on_release=lambda x: diag.dismiss()
                ),
                MDButton(
                    MDButtonText(text="Save"),
                    style="text",
                    on_release=lambda x: self.save_file(file_name_input.text) or diag.dismiss()
                ),
                spacing="8dp",
            ),
            # -------------------------------------------------------------
            auto_dismiss=False,
        )
        diag.open()

    def preview_tree(self):
        self.preview.text = self.editor.text


class LLMGeneratorScreen(MDScreen):
    prompt_input: MDTextField = None
    input_tree: MDTextField = None
    output_tree: MDTextField = None
    logs_text: MDTextField = None
    send_btn: MDButton = None

    def send_to_llm(self):
        url = settings.data.get('llm_api_url', LLM_API_URL_PLACEHOLDER)
        if url == LLM_API_URL_PLACEHOLDER:
            show_error('Please set the LLM API URL in Settings.')
            return
        payload = {
            'Tree': self.input_tree.text or '',
            'Prompt': self.prompt_input.text or '',
        }
        self.send_btn.disabled = True
        threading.Thread(target=self._post_thread, args=(url, payload), daemon=True).start()

    def _post_thread(self, url, payload):
        try:
            resp = requests.post(url, json=payload, timeout=60)
            resp.raise_for_status()
            data = resp.json()
            tree = data.get('Tree', '')
            changes = data.get('changes', {})
            self._on_llm_success(tree, changes)
        except Exception as e:
            self._on_llm_error(str(e))

    @mainthread
    def _on_llm_success(self, tree, changes):
        self.output_tree.text = tree
        lines = []
        for t in ('added', 'edited', 'deleted'):
            items = changes.get(t, [])
            if items:
                lines.append(f"== {t.upper()} ==")
                for it in items:
                    lines.append(str(it))
        self.logs_text.text = '\n'.join(lines)
        self.send_btn.disabled = False

    @mainthread
    def _on_llm_error(self, err):
        self.send_btn.disabled = False
        show_error(f'LLM request failed:\n{err}')

class ReverseScreen(MDScreen):
    chooser = ObjectProperty(None)
    generated = ObjectProperty(None)

    def pick_dir_and_reverse(self):
        path = self.chooser.path
        if not path:
            show_error('Select a directory first')
            return
        threading.Thread(target=self._reverse_thread, args=(path,), daemon=True).start()

    def _reverse_thread(self, path):
        # This assumes mktree is available as a Python module in PATH or via alias
        # We'll call: python -m mktree <dir> --reverse --no-content
        try:
            cmd = ['python', '-m', 'mktree', path, '--reverse', '--no-content']
            proc = subprocess.run(cmd, capture_output=True, text=True, check=False)
            if proc.returncode != 0:
                out = proc.stderr or proc.stdout
                self._on_reverse_error(out)
            else:
                out = proc.stdout
                self._on_reverse_success(out)
        except Exception as e:
            self._on_reverse_error(str(e))

    @mainthread
    def _on_reverse_success(self, text):
        self.generated.text = text

    @mainthread
    def _on_reverse_error(self, text):
        show_error(f'Reverse failed:\n{text}')

class TemplatesScreen(MDScreen):
    templates_list = ObjectProperty(None)
    template_preview = ObjectProperty(None)

    def load_templates(self):
        tdir = os.path.join(project_path, 'templates')
        items = []
        if os.path.exists(tdir):
            for p in Path(tdir).glob('*.tree'):
                items.append(str(p.name))
        self.templates_list.data = [
            {
                "text": name,
                "on_release": lambda btn_name=name: self.on_template_select(btn_name)
            }
            for name in items
        ]

    def on_template_select(self, name):
        path = os.path.join(project_path, 'templates', name)
        if os.path.exists(path):
            self.template_preview.text = Path(path).read_text(encoding='utf-8')


class SettingsScreen(MDScreen):
    url_input_text = StringProperty("")
    preview_colored = BooleanProperty(True)
    editor_font_size = NumericProperty(14)

    def on_kv_post(self, base_widget):
        app = MDApp.get_running_app()
        self.url_input_text = app.settings.data.get("llm_api_url", "")
        self.preview_colored = app.settings.data.get("preview_colored", True)
        self.editor_font_size = app.settings.data.get("editor_font_size", 14)

    def toggle_preview(self):
        self.preview_colored = not self.preview_colored

    def save_settings(self):
        settings.data['llm_api_url'] = self.url_input_text
        settings.data['preview_colored'] = self.preview_colored
        try:
            settings.data['editor_font_size'] = int(self.editor_font_size)
        except Exception:
            pass
        settings.save()
        MDSnackbar(
            MDSnackbarText(
                text="Settings saved!",
            ),
            y=dp(24),
            pos_hint={"center_x": 0.5},
            size_hint_x=0.8,
            duration=1.0 
        ).open()
       

class MyScreenManager(MDScreenManager):
    current_file = ""

class MkTree(MDApp):
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.settings = settings
        Window.bind(on_keyboard=self.events)
        self.manager_open = False
        self.file_manager = MDFileManager(
            exit_manager=self.exit_manager,
            select_path=self.select_path,
            ext = [".tree"],
        )


    def build(self):
        self.title = "mktree GUI"
        
        self.theme_cls.theme_style = "Dark"  # or "Light"
        self.theme_cls.primary_palette = "Blue"
        return Builder.load_file('mktree_gui.kv')
    def on_stop(self):
        settings.save()
    def file_manager_open(self):
        self.file_manager.show(os.path.abspath("."))
        self.manager_open = True

    def select_path(self, path: str):
        '''
        It will be called when you click on the file name
        or the catalog selection button.

        :param path: path to the selected directory or file;
        '''

        self.exit_manager()
        MDSnackbar(
            MDSnackbarText(
                text=path,
            ),
            y=dp(24),
            pos_hint={"center_x": 0.5},
            size_hint_x=0.8,
            duration=1.0 
        ).open()
        editor_screen = MDApp.get_running_app().root.screen_manager.get_screen('editor')
        editor_screen.editor.text = Path(path).read_text(encoding='utf-8')
        MDApp.get_running_app().root.screen_manager.current = "editor"
    def exit_manager(self, *args):
        '''Called when the user reaches the root of the directory tree.'''
        self.manager_open = False
        self.file_manager.close()

    def events(self, instance, keyboard, keycode, text, modifiers):
        '''Called when buttons are pressed on the mobile device.'''

        if keyboard in (1001, 27):
            if self.manager_open:
                self.file_manager.back()
        return True
if __name__ == '__main__':
    MkTree().run()