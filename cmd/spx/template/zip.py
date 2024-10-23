import os
import zipfile
import subprocess
import os

def zipdir(path, ziph, skip_dirs):
    # ziph is zipfile handle
    for root, dirs, files in os.walk(path):
        if root == path:  # Apply filter only at top level
            dirs[:] = [d for d in dirs if d not in skip_dirs]
            files[:] = [f for f in files if f not in skip_dirs]
        for file in files:
            arcname = os.path.join('demo', os.path.relpath(os.path.join(root, file), path))
            ziph.write(os.path.join(root, file), arcname=arcname)

skip_dirs = ["lib", ".godot", ".builds","zip.py"]
zipf = zipfile.ZipFile('./.builds/web/game2.zip', 'w', zipfile.ZIP_STORED)
zipdir('.', zipf, skip_dirs)
zipf.close()

os.chdir("./.builds/web")
import shutil
winrar_path = "WinRAR.exe"
subprocess.run([winrar_path, "x", "./game2.zip", "./"])
subprocess.run([winrar_path, "a", "-afzip", "-ep1", "./game.zip", "./demo"])
os.remove("./game2.zip")
shutil.rmtree("./demo")