import yt_dlp
import os

def download_video(url, output_path="downloads", quality="best"):
    """
    Download a YouTube video
    
    Args:
        url: YouTube video URL
        output_path: Directory to save the video
        quality: 'best', 'worst', or specific resolution like '720p'
    """
    # Create output directory if it doesn't exist
    if not os.path.exists(output_path):
        os.makedirs(output_path)
    
    # Configure download options
    ydl_opts = {
        'format': 'bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best',
        'outtmpl': os.path.join(output_path, '%(title)s.%(ext)s'),
        'quiet': False,
        'no_warnings': False,
        'noplaylist': True,  # Only download single video, not playlist
    }
    
    # Adjust format based on quality preference
    if quality == 'worst':
        ydl_opts['format'] = 'worst[ext=mp4]/worst'
    elif quality != 'best':
        ydl_opts['format'] = f'bestvideo[height<={quality[:-1]}][ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best'
    
    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            print(f"Downloading: {url}")
            info = ydl.extract_info(url, download=True)
            print(f"✓ Downloaded: {info['title']}")
            return True
    except Exception as e:
        print(f"✗ Error downloading video: {e}")
        return False

def download_audio_only(url, output_path="downloads"):
    """
    Download only the audio from a YouTube video
    """
    if not os.path.exists(output_path):
        os.makedirs(output_path)
    
    ydl_opts = {
        'format': 'bestaudio/best',
        'outtmpl': os.path.join(output_path, '%(title)s.%(ext)s'),
        'postprocessors': [{
            'key': 'FFmpegExtractAudio',
            'preferredcodec': 'mp3',
            'preferredquality': '192',
        }],
        'quiet': False,
    }
    
    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            print(f"Downloading audio: {url}")
            info = ydl.extract_info(url, download=True)
            print(f"✓ Downloaded audio: {info['title']}")
            return True
    except Exception as e:
        print(f"✗ Error downloading audio: {e}")
        return False

def get_video_info(url):
    """
    Get information about a YouTube video without downloading
    """
    ydl_opts = {'quiet': True}
    
    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)
            print(f"\nTitle: {info['title']}")
            print(f"Duration: {info['duration']} seconds")
            print(f"Views: {info['view_count']:,}")
            print(f"Uploader: {info['uploader']}")
            print(f"Upload Date: {info['upload_date']}")
            return info
    except Exception as e:
        print(f"✗ Error getting video info: {e}")
        return None

# Example usage
if __name__ == "__main__":
    print("=== YouTube Downloader ===\n")
    
    # Example: Download a video
    video_url = input("Enter YouTube URL: ").strip()
    
    if video_url:
        print("\nOptions:")
        print("1. Download video (best quality)")
        print("2. Download video (720p)")
        print("3. Download audio only (MP3)")
        print("4. Get video info only")
        
        choice = input("\nSelect option (1-4): ").strip()
        
        if choice == "1":
            download_video(video_url)
        elif choice == "2":
            download_video(video_url, quality="720p")
        elif choice == "3":
            download_audio_only(video_url)
        elif choice == "4":
            get_video_info(video_url)
        else:
            print("Invalid option")
    else:
        print("No URL provided")