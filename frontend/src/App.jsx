import { useState, useEffect } from 'react';
import {
  Download,
  Music,
  Video,
  CheckCircle,
  AlertCircle,
  Loader,
  RefreshCw,
  Settings,
  FileVideo,
  FileAudio,
} from 'lucide-react';

const API_URL = 'http://localhost:8080/api';

function App() {
  const [url, setUrl] = useState('');
  const [format, setFormat] = useState('video');
  const [quality, setQuality] = useState('best');
  const [extension, setExtension] = useState('mp4');
  const [downloads, setDownloads] = useState([]);
  const [formats, setFormats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [detectedPlatform, setDetectedPlatform] = useState('');

  useEffect(() => {
    fetchFormats();
    fetchDownloads();
    const interval = setInterval(fetchDownloads, 3000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    // Detect platform when URL changes
    if (url) {
      if (url.includes('tiktok.com') || url.includes('vm.tiktok.com')) {
        setDetectedPlatform('TikTok');
      } else if (url.includes('youtube.com') || url.includes('youtu.be')) {
        setDetectedPlatform('YouTube');
      } else {
        setDetectedPlatform('');
      }
    } else {
      setDetectedPlatform('');
    }
  }, [url]);

  const fetchFormats = async () => {
    try {
      const response = await fetch(`${API_URL}/formats`);
      const data = await response.json();
      setFormats(data);
    } catch (error) {
      console.error('Error fetching formats:', error);
    }
  };

  const fetchDownloads = async () => {
    try {
      const response = await fetch(`${API_URL}/downloads`);
      const data = await response.json();
      setDownloads(data || []);
    } catch (error) {
      console.error('Error fetching downloads:', error);
    }
  };

  const handleFormatChange = (newFormat) => {
    setFormat(newFormat);
    setQuality('best');
    setExtension(newFormat === 'audio' ? 'mp3' : 'mp4');
  };

  const handleSubmit = async () => {
    if (!url) return;

    setLoading(true);
    setMessage('');

    try {
      const response = await fetch(`${API_URL}/download`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, format, quality, extension }),
      });

      if (response.ok) {
        const data = await response.json();
        setMessage(`Download started successfully! Platform: ${data.platform}`);
        setUrl('');
        setDetectedPlatform('');
        fetchDownloads();
      } else {
        setMessage('Failed to start download');
      }
    } catch (error) {
      setMessage('Error: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'failed':
        return <AlertCircle className="w-5 h-5 text-red-500" />;
      case 'processing':
      case 'streaming':
        return <Loader className="w-5 h-5 text-blue-500 animate-spin" />;
      default:
        return <RefreshCw className="w-5 h-5 text-gray-400" />;
    }
  };

  const getPlatformBadge = (platform) => {
    if (platform === 'tiktok') {
      return (
        <span className="px-3 py-1.5 bg-gradient-to-r from-cyan-100 to-pink-100 text-pink-700 rounded-full font-bold text-xs">
          TikTok
        </span>
      );
    }
    return (
      <span className="px-3 py-1.5 bg-red-100 text-red-700 rounded-full font-bold text-xs">
        YouTube
      </span>
    );
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  const handleDownloadFile = async (id, title, ext) => {
    try {
      const link = document.createElement('a');
      link.href = `${API_URL}/stream/${id}`;
      link.download = `${title || 'download'}.${ext}`;
      link.target = '_blank';
      
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      setMessage('Download started! Check your browser downloads.');
    } catch (error) {
      console.error('Download error:', error);
      setMessage('Error downloading file: ' + error.message);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-100 via-purple-50 to-pink-100 p-4 sm:p-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-5xl font-bold text-gray-800 mb-3 flex items-center justify-center gap-3">
            <Download className="w-12 h-12 text-purple-600" />
            Universal Downloader Pro
          </h1>
          <p className="text-gray-600 text-lg">Download from YouTube & TikTok with custom quality settings</p>
          <div className="flex items-center justify-center gap-4 mt-4">
            <span className="px-4 py-2 bg-red-100 text-red-700 rounded-full font-semibold text-sm">
              YouTube
            </span>
            <span className="px-4 py-2 bg-gradient-to-r from-cyan-100 to-pink-100 text-pink-700 rounded-full font-bold text-sm">
              TikTok
            </span>
          </div>
        </div>

        <div className="bg-white rounded-3xl shadow-2xl p-6 sm:p-8 mb-8">
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Video URL {detectedPlatform && (
                  <span className="ml-2 px-3 py-1 bg-green-100 text-green-700 rounded-full text-xs font-bold">
                    ✓ {detectedPlatform} detected
                  </span>
                )}
              </label>
              <input
                type="text"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSubmit()}
                placeholder="Paste YouTube or TikTok URL here..."
                className="w-full px-5 py-4 border-2 border-gray-200 rounded-xl focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none transition text-lg"
              />
              <p className="text-xs text-gray-500 mt-2">
                Supports: youtube.com, youtu.be, tiktok.com, vm.tiktok.com
              </p>
            </div>

            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-3">
                Download Type
              </label>
              <div className="grid grid-cols-2 gap-4">
                <button
                  type="button"
                  onClick={() => handleFormatChange('video')}
                  className={`p-5 rounded-xl border-2 transition-all transform hover:scale-105 ${
                    format === 'video'
                      ? 'border-purple-500 bg-gradient-to-br from-purple-50 to-blue-50 shadow-lg'
                      : 'border-gray-200 hover:border-gray-300 bg-white'
                  }`}
                >
                  <Video className={`w-8 h-8 mx-auto mb-2 ${format === 'video' ? 'text-purple-600' : 'text-gray-400'}`} />
                  <div className="font-semibold text-lg">Video</div>
                  <div className="text-sm text-gray-500">MP4, WebM, MKV</div>
                </button>
                <button
                  type="button"
                  onClick={() => handleFormatChange('audio')}
                  className={`p-5 rounded-xl border-2 transition-all transform hover:scale-105 ${
                    format === 'audio'
                      ? 'border-purple-500 bg-gradient-to-br from-purple-50 to-pink-50 shadow-lg'
                      : 'border-gray-200 hover:border-gray-300 bg-white'
                  }`}
                >
                  <Music className={`w-8 h-8 mx-auto mb-2 ${format === 'audio' ? 'text-purple-600' : 'text-gray-400'}`} />
                  <div className="font-semibold text-lg">Audio</div>
                  <div className="text-sm text-gray-500">MP3, AAC, Opus</div>
                </button>
              </div>
            </div>

            <button
              type="button"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="flex items-center gap-2 text-purple-600 hover:text-purple-700 font-medium transition"
            >
              <Settings className="w-5 h-5" />
              {showAdvanced ? 'Hide' : 'Show'} Advanced Settings
            </button>

            {showAdvanced && formats && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 p-4 bg-gray-50 rounded-xl border border-gray-200">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    Quality
                  </label>
                  <select
                    value={quality}
                    onChange={(e) => setQuality(e.target.value)}
                    className="w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none transition bg-white"
                  >
                    {(format === 'video' ? formats.video_qualities : formats.audio_qualities).map((q) => (
                      <option key={q} value={q}>
                        {q === 'best' ? 'Best Available' : q}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    Format
                  </label>
                  <select
                    value={extension}
                    onChange={(e) => setExtension(e.target.value)}
                    className="w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none transition bg-white"
                  >
                    {(format === 'video' ? formats.video_formats : formats.audio_formats).map((f) => (
                      <option key={f} value={f}>
                        {f.toUpperCase()}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            )}

            <button
              onClick={handleSubmit}
              disabled={loading || !url}
              className="w-full bg-gradient-to-r from-purple-600 via-purple-500 to-blue-500 text-white py-4 rounded-xl font-semibold text-lg hover:from-purple-700 hover:via-purple-600 hover:to-blue-600 transition-all transform hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none flex items-center justify-center gap-3 shadow-lg"
            >
              {loading ? (
                <>
                  <Loader className="w-6 h-6 animate-spin" />
                  Processing Download...
                </>
              ) : (
                <>
                  <Download className="w-6 h-6" />
                  Start Download
                </>
              )}
            </button>

            {message && (
              <div
                className={`p-4 rounded-xl font-medium ${
                  message.includes('success') || message.includes('started')
                    ? 'bg-green-100 text-green-800 border border-green-200'
                    : 'bg-red-100 text-red-800 border border-red-200'
                }`}
              >
                {message}
              </div>
            )}
          </div>
        </div>

        <div className="bg-white rounded-3xl shadow-2xl p-6 sm:p-8">
          <h2 className="text-3xl font-bold text-gray-800 mb-6 flex items-center gap-3">
            <RefreshCw className="w-8 h-8 text-purple-600" />
            Download History
          </h2>

          <div className="space-y-4">
            {downloads.length === 0 ? (
              <div className="text-center py-16">
                <div className="inline-block p-6 bg-gray-100 rounded-full mb-4">
                  <Download className="w-12 h-12 text-gray-400" />
                </div>
                <p className="text-gray-500 text-lg">No downloads yet</p>
                <p className="text-gray-400 text-sm mt-2">Start by adding a YouTube or TikTok URL above</p>
              </div>
            ) : (
              downloads.map((download) => (
                <div
                  key={download.id}
                  className="border-2 border-gray-100 rounded-xl p-5 hover:shadow-xl hover:border-purple-200 transition-all bg-gradient-to-r from-white to-gray-50"
                >
                  <div className="flex items-start justify-between gap-4 flex-wrap">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-3">
                        {getStatusIcon(download.status)}
                        <h3 className="font-semibold text-gray-800 text-lg truncate">
                          {download.title || 'Processing...'}
                        </h3>
                      </div>
                      <p className="text-sm text-gray-500 truncate mb-3">
                        {download.url}
                      </p>
                      <div className="flex flex-wrap items-center gap-3 text-xs">
                        {download.platform && getPlatformBadge(download.platform)}
                        <span className="flex items-center gap-1.5 px-3 py-1.5 bg-purple-100 text-purple-700 rounded-full font-medium">
                          {download.format === 'audio' ? (
                            <FileAudio className="w-3.5 h-3.5" />
                          ) : (
                            <FileVideo className="w-3.5 h-3.5" />
                          )}
                          {download.extension.toUpperCase()}
                        </span>
                        <span className="px-3 py-1.5 bg-blue-100 text-blue-700 rounded-full font-medium">
                          {download.quality}
                        </span>
                        <span className={`px-3 py-1.5 rounded-full font-medium capitalize ${
                          download.status === 'completed' ? 'bg-green-100 text-green-700' :
                          download.status === 'processing' || download.status === 'streaming' ? 'bg-blue-100 text-blue-700' :
                          download.status === 'failed' ? 'bg-red-100 text-red-700' :
                          'bg-gray-100 text-gray-700'
                        }`}>
                          {download.status}
                        </span>
                        {download.duration > 0 && (
                          <span className="px-3 py-1.5 bg-gray-100 text-gray-700 rounded-full font-medium">
                            ⏱ {formatDuration(download.duration)}
                          </span>
                        )}
                        {download.file_size > 0 && (
                          <span className="px-3 py-1.5 bg-gray-100 text-gray-700 rounded-full font-medium">
                            📦 {formatFileSize(download.file_size)}
                          </span>
                        )}
                        <span className="text-gray-400">
                          {formatDate(download.created_at)}
                        </span>
                      </div>
                    </div>

                    {(download.status === 'ready' || download.status === 'completed') && (
                      <button
                        onClick={() =>
                          handleDownloadFile(
                            download.id,
                            download.title,
                            download.extension
                          )
                        }
                        className="px-6 py-3 bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-lg hover:from-purple-700 hover:to-blue-700 transition-all transform hover:scale-105 font-semibold flex items-center gap-2 whitespace-nowrap shadow-lg"
                      >
                        <Download className="w-5 h-5" />
                        Download
                      </button>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;