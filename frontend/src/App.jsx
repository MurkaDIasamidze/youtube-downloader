import React, { useState, useEffect } from 'react';
import { Download, Music, Video, CheckCircle, AlertCircle, Loader, RefreshCw } from 'lucide-react';

const API_URL = 'http://localhost:8080/api';

export default function App() {
  const [url, setUrl] = useState('');
  const [format, setFormat] = useState('video');
  const [downloads, setDownloads] = useState([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    fetchDownloads();
    const interval = setInterval(fetchDownloads, 3000);
    return () => clearInterval(interval);
  }, []);

  const fetchDownloads = async () => {
    try {
      const response = await fetch(`${API_URL}/downloads`);
      const data = await response.json();
      setDownloads(data || []);
    } catch (error) {
      console.error('Error fetching downloads:', error);
    }
  };

  const handleSubmit = async () => {
    if (!url) return;

    setLoading(true);
    setMessage('');

    try {
      const response = await fetch(`${API_URL}/download`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, format }),
      });

      if (response.ok) {
        setMessage('Download started successfully!');
        setUrl('');
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
        return <Loader className="w-5 h-5 text-blue-500 animate-spin" />;
      default:
        return <RefreshCw className="w-5 h-5 text-gray-400" />;
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const formatDuration = (seconds) => {
    if (!seconds) return '';
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${minutes}m ${secs}s`;
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-blue-50 p-8">
      <div className="max-w-6xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2 flex items-center justify-center gap-3">
            <Download className="w-10 h-10 text-purple-600" />
            YouTube Downloader
          </h1>
          <p className="text-gray-600">Download videos and audio from YouTube</p>
        </div>

        <div className="bg-white rounded-2xl shadow-xl p-8 mb-8">
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                YouTube URL
              </label>
              <input
                type="text"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSubmit()}
                placeholder="https://www.youtube.com/watch?v=..."
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none transition"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-3">
                Download Format
              </label>
              <div className="flex gap-4">
                <button
                  type="button"
                  onClick={() => setFormat('video')}
                  className={`flex-1 p-4 rounded-lg border-2 transition ${
                    format === 'video'
                      ? 'border-purple-500 bg-purple-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <Video className="w-6 h-6 mx-auto mb-2 text-purple-600" />
                  <div className="font-medium">Video</div>
                  <div className="text-sm text-gray-500">MP4 Format</div>
                </button>
                <button
                  type="button"
                  onClick={() => setFormat('audio')}
                  className={`flex-1 p-4 rounded-lg border-2 transition ${
                    format === 'audio'
                      ? 'border-purple-500 bg-purple-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <Music className="w-6 h-6 mx-auto mb-2 text-purple-600" />
                  <div className="font-medium">Audio</div>
                  <div className="text-sm text-gray-500">MP3 Format</div>
                </button>
              </div>
            </div>

            <button
              onClick={handleSubmit}
              disabled={loading || !url}
              className="w-full bg-gradient-to-r from-purple-600 to-blue-600 text-white py-3 rounded-lg font-medium hover:from-purple-700 hover:to-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <Loader className="w-5 h-5 animate-spin" />
                  Starting Download...
                </>
              ) : (
                <>
                  <Download className="w-5 h-5" />
                  Start Download
                </>
              )}
            </button>

            {message && (
              <div className={`p-4 rounded-lg ${
                message.includes('success') ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
              }`}>
                {message}
              </div>
            )}
          </div>
        </div>

        <div className="bg-white rounded-2xl shadow-xl p-8">
          <h2 className="text-2xl font-bold text-gray-800 mb-6 flex items-center gap-2">
            <RefreshCw className="w-6 h-6 text-purple-600" />
            Download History
          </h2>

          <div className="space-y-4">
            {downloads.length === 0 ? (
              <div className="text-center py-12 text-gray-500">
                No downloads yet. Start by adding a YouTube URL above.
              </div>
            ) : (
              downloads.map((download) => (
                <div
                  key={download.id}
                  className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-2">
                        {getStatusIcon(download.status)}
                        <h3 className="font-medium text-gray-800 truncate">
                          {download.title || 'Processing...'}
                        </h3>
                      </div>
                      <p className="text-sm text-gray-500 truncate mb-2">
                        {download.url}
                      </p>
                      <div className="flex items-center gap-4 text-xs text-gray-400">
                        <span className="flex items-center gap-1">
                          {download.format === 'audio' ? (
                            <Music className="w-3 h-3" />
                          ) : (
                            <Video className="w-3 h-3" />
                          )}
                          {download.format.toUpperCase()}
                        </span>
                        <span>{formatDate(download.created_at)}</span>
                        <span className="capitalize">{download.status}</span>
                        {download.duration > 0 && (
                          <span className="text-green-600 font-medium">
                            ⏱ {formatDuration(download.duration)}
                          </span>
                        )}
                      </div>
                    </div>

                    {download.status === 'completed' && (
                      <div className="flex gap-2">
                        <a
                          href={`${API_URL}/file/${download.id}`}
                          download={download.title ? `${download.title}.${download.format === 'audio' ? 'mp3' : 'mp4'}` : undefined}
                          className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition text-sm font-medium flex items-center gap-2 whitespace-nowrap"
                        >
                          <Download className="w-4 h-4" />
                          Download
                        </a>
                      </div>
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