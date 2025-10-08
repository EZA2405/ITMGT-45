from flask import Flask, render_template, request, jsonify
import threading, os, yt_dlp, uuid

app = Flask(__name__)
downloads_progress = {}

def download_with_progress(url, folder, mode, download_id):
    os.makedirs(f"downloads/{folder}", exist_ok=True)

    def hook(d):
        if d['status'] == 'downloading':
            total = d.get('total_bytes') or d.get('total_bytes_estimate', 1)
            downloaded = d.get('downloaded_bytes', 0)
            percent = int(downloaded / total * 100)
            downloads_progress[download_id] = {
                "status": f"Downloading... {percent}%",
                "progress": percent
            }
        elif d['status'] == 'finished':
            downloads_progress[download_id] = {"status": "Processing...", "progress": 100}

    opts = {
        'outtmpl': os.path.join('downloads', folder, '%(title)s.%(ext)s'),
        'progress_hooks': [hook],
    }

    if mode == "audio":
        opts.update({
            'format': 'bestaudio/best',
            'postprocessors': [{
                'key': 'FFmpegExtractAudio',
                'preferredcodec': 'mp3',
                'preferredquality': '192',
            }]
        })
    else:
        opts.update({'format': 'bestvideo+bestaudio/best'})

    with yt_dlp.YoutubeDL(opts) as ydl:
        ydl.download([url])

    downloads_progress[download_id] = {"status": "✅ Done!", "progress": 100}

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/download', methods=['POST'])
def start_download():
    data = request.get_json()
    url, folder, mode = data['url'], data['folder'], data['mode']
    download_id = str(uuid.uuid4())
    downloads_progress[download_id] = {"status": "Starting...", "progress": 0}

    thread = threading.Thread(target=download_with_progress, args=(url, folder, mode, download_id))
    thread.start()

    return jsonify({"id": download_id})

@app.route('/progress/<download_id>')
def get_progress(download_id):
    return jsonify(downloads_progress.get(download_id, {"status": "Unknown ID", "progress": 0}))

if __name__ == "__main__":
    app.run(debug=True)
