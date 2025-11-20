import React, { useState, useEffect } from 'react';
import {
  StyleSheet,
  Text,
  View,
  TextInput,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
  Modal,
  Linking,
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { BlurView } from 'expo-blur';
import * as DocumentPicker from 'expo-document-picker';
import { Upload, File, X, CheckCircle, AlertCircle, Settings, Clock, Trash2, Copy, ExternalLink, ChevronDown } from 'lucide-react-native';
import { saveServerUrl, getServerUrl, saveHistoryItem, getHistory, clearHistory, HistoryItem } from './storage';
import * as Clipboard from 'expo-clipboard';

interface FileItem {
  uri: string;
  name: string;
  size?: number;
  mimeType?: string;
  status: 'pending' | 'uploading' | 'success' | 'error';
  progress: number;
  downloadUrl?: string;
}

const TTL_OPTIONS = [
  { label: '1 Hour', value: '3600' },
  { label: '1 Day', value: '86400' },
  { label: '1 Week', value: '604800' },
  { label: '1 Month', value: '2592000' },
];

export default function App() {
  const [serverUrl, setServerUrl] = useState('https://temp.low-stack.tech');
  const [ttl, setTtl] = useState('3600');
  const [files, setFiles] = useState<FileItem[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [showTTLSelector, setShowTTLSelector] = useState(false);

  useEffect(() => {
    loadSettings();
    loadHistory();
  }, []);

  const loadSettings = async () => {
    const url = await getServerUrl();
    setServerUrl(url);
  };

  const loadHistory = async () => {
    const items = await getHistory();
    setHistory(items);
  };

  const handleServerUrlChange = (url: string) => {
    setServerUrl(url);
    saveServerUrl(url);
  };

  const handleClearHistory = async () => {
    Alert.alert(
      'Clear History',
      'Are you sure you want to clear your upload history?',
      [
        { text: 'Cancel', style: 'cancel' },
        { 
          text: 'Clear', 
          style: 'destructive', 
          onPress: async () => {
            await clearHistory();
            setHistory([]);
          } 
        }
      ]
    );
  };

  const pickFiles = async () => {
    try {
      const result = await DocumentPicker.getDocumentAsync({
        type: '*/*',
        multiple: true,
        copyToCacheDirectory: true,
      });

      if (!result.canceled) {
        const newFiles = result.assets.map((asset) => ({
          uri: asset.uri,
          name: asset.name,
          size: asset.size,
          mimeType: asset.mimeType,
          status: 'pending' as const,
          progress: 0,
        }));
        setFiles((prev) => [...prev, ...newFiles]);
      }
    } catch (err) {
      Alert.alert('Error', 'Failed to pick files');
    }
  };

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const copyToClipboard = async (text: string) => {
    await Clipboard.setStringAsync(text);
    Alert.alert('Copied', 'Link copied to clipboard');
  };

  const uploadFiles = async () => {
    if (files.length === 0) {
      Alert.alert('Error', 'No files selected');
      return;
    }

    setIsUploading(true);

    // Create FormData
    const formData = new FormData();
    formData.append('ttl_seconds', ttl);

    files.forEach((file) => {
      // @ts-ignore: React Native FormData expects object with uri, name, type
      formData.append('file', {
        uri: file.uri,
        name: file.name,
        type: file.mimeType || 'application/octet-stream',
      });
    });

    try {
      // Update all files to uploading state
      setFiles((prev) =>
        prev.map((f) => ({ ...f, status: 'uploading', progress: 0 }))
      );

      const xhr = new XMLHttpRequest();
      
      xhr.open('POST', `${serverUrl}/api/upload`);
      
      // Track progress
      if (xhr.upload) {
        xhr.upload.onprogress = (event) => {
          if (event.lengthComputable) {
            const progress = event.loaded / event.total;
            setFiles((prev) =>
              prev.map((f) => ({ ...f, progress }))
            );
          }
        };
      }

      xhr.onload = async () => {
        if (xhr.status === 200) {
          const response = JSON.parse(xhr.responseText);
          if (response.success) {
             setFiles((prev) =>
              prev.map((f) => {
                const uploadedFile = response.files.find((uf: any) => uf.filename === f.name);
                return {
                  ...f,
                  status: 'success',
                  progress: 1,
                  downloadUrl: uploadedFile?.download_url,
                };
              })
            );
            
            // Save to history
            for (const file of response.files) {
              const historyItem: HistoryItem = {
                id: file.id,
                filename: file.filename,
                size: file.size,
                downloadUrl: file.download_url,
                uploadedAt: Date.now(),
              };
              await saveHistoryItem(historyItem);
            }
            await loadHistory();
            
            Alert.alert('Success', 'Files uploaded successfully!');
          } else {
             setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
             Alert.alert('Upload Failed', response.error || 'Unknown error');
          }
        } else {
          setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
          Alert.alert('Error', `Server returned status ${xhr.status}`);
        }
        setIsUploading(false);
      };

      xhr.onerror = () => {
        setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
        Alert.alert('Error', 'Network request failed');
        setIsUploading(false);
      };

      xhr.send(formData);

    } catch (error) {
      setFiles((prev) => prev.map((f) => ({ ...f, status: 'error' })));
      Alert.alert('Error', 'An unexpected error occurred');
      setIsUploading(false);
    }
  };

  const getTTLLabel = () => {
    const option = TTL_OPTIONS.find(opt => opt.value === ttl);
    return option ? option.label : 'Custom';
  };

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a1a1a', '#2d2d2d', '#1a1a1a']}
        style={StyleSheet.absoluteFill}
      />
      
      <View style={styles.headerContainer}>
        <Text style={styles.title}>Temp Upload</Text>
        <TouchableOpacity onPress={() => setShowSettings(true)} style={styles.settingsButton}>
          <Settings color="#fff" size={24} />
        </TouchableOpacity>
      </View>

      <ScrollView contentContainerStyle={styles.scrollContent}>
        <View style={styles.card}>
          <Text style={styles.label}>File Expiration</Text>
          <TouchableOpacity 
            style={styles.selectorButton}
            onPress={() => setShowTTLSelector(true)}
          >
            <Text style={styles.selectorText}>{getTTLLabel()}</Text>
            <ChevronDown color="#aaa" size={20} />
          </TouchableOpacity>
        </View>

        <TouchableOpacity
          style={styles.pickButton}
          onPress={pickFiles}
          disabled={isUploading}
        >
          <LinearGradient
            colors={['#4c669f', '#3b5998', '#192f6a']}
            style={styles.gradientButton}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
          >
            <File color="white" size={24} />
            <Text style={styles.buttonText}>Select Files</Text>
          </LinearGradient>
        </TouchableOpacity>

        {files.length > 0 && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Current Uploads</Text>
            {files.map((file, index) => (
              <BlurView intensity={20} tint="dark" key={index} style={styles.fileItem}>
                <View style={styles.fileInfo}>
                  <Text style={styles.fileName} numberOfLines={1}>
                    {file.name}
                  </Text>
                  <Text style={styles.fileSize}>
                    {file.size ? (file.size / 1024).toFixed(1) + ' KB' : 'Unknown size'}
                  </Text>
                  {file.status === 'uploading' && (
                    <View style={styles.progressBar}>
                      <View 
                        style={[
                          styles.progressFill, 
                          { width: `${file.progress * 100}%` }
                        ]} 
                      />
                    </View>
                  )}
                </View>
                <View style={styles.fileActions}>
                  {file.status === 'success' ? (
                    <TouchableOpacity onPress={() => file.downloadUrl && copyToClipboard(file.downloadUrl)}>
                       <CheckCircle color="#4caf50" size={20} />
                    </TouchableOpacity>
                  ) : file.status === 'error' ? (
                    <AlertCircle color="#f44336" size={20} />
                  ) : (
                    !isUploading && (
                      <TouchableOpacity onPress={() => removeFile(index)}>
                        <X color="#ff5252" size={20} />
                      </TouchableOpacity>
                    )
                  )}
                </View>
              </BlurView>
            ))}
            
            <TouchableOpacity
              style={[styles.uploadButton, isUploading && styles.disabledButton]}
              onPress={uploadFiles}
              disabled={isUploading}
            >
              {isUploading ? (
                <ActivityIndicator color="white" />
              ) : (
                <>
                  <Upload color="white" size={24} style={{ marginRight: 8 }} />
                  <Text style={styles.uploadButtonText}>Upload {files.length} Files</Text>
                </>
              )}
            </TouchableOpacity>
          </View>
        )}

        {history.length > 0 && (
          <View style={styles.section}>
            <View style={styles.sectionHeader}>
              <Clock color="#aaa" size={20} />
              <Text style={styles.sectionTitle}>History</Text>
            </View>
            {history.map((item) => (
              <BlurView intensity={10} tint="dark" key={item.id} style={styles.historyItem}>
                <View style={styles.fileInfo}>
                  <Text style={styles.fileName} numberOfLines={1}>{item.filename}</Text>
                  <Text style={styles.fileSize}>
                    {(item.size / 1024).toFixed(1)} KB • {new Date(item.uploadedAt).toLocaleDateString()}
                  </Text>
                </View>
                <View style={styles.historyActions}>
                  <TouchableOpacity onPress={() => copyToClipboard(item.downloadUrl)} style={styles.actionButton}>
                    <Copy color="#fff" size={18} />
                  </TouchableOpacity>
                  <TouchableOpacity onPress={() => Linking.openURL(item.downloadUrl)} style={styles.actionButton}>
                    <ExternalLink color="#fff" size={18} />
                  </TouchableOpacity>
                </View>
              </BlurView>
            ))}
          </View>
        )}
      </ScrollView>

      {/* Settings Modal */}
      <Modal
        visible={showSettings}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setShowSettings(false)}
      >
        <BlurView intensity={50} tint="dark" style={styles.modalContainer}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>Settings</Text>
              <TouchableOpacity onPress={() => setShowSettings(false)}>
                <X color="#fff" size={24} />
              </TouchableOpacity>
            </View>

            <Text style={styles.label}>Server URL</Text>
            <TextInput
              style={styles.input}
              value={serverUrl}
              onChangeText={handleServerUrlChange}
              placeholder="https://temp.low-stack.tech"
              placeholderTextColor="#666"
              autoCapitalize="none"
            />

            <TouchableOpacity style={styles.clearButton} onPress={handleClearHistory}>
              <Trash2 color="#ff5252" size={20} style={{ marginRight: 8 }} />
              <Text style={styles.clearButtonText}>Clear History</Text>
            </TouchableOpacity>
          </View>
        </BlurView>
      </Modal>

      {/* TTL Selector Modal */}
      <Modal
        visible={showTTLSelector}
        animationType="fade"
        transparent={true}
        onRequestClose={() => setShowTTLSelector(false)}
      >
        <TouchableOpacity 
          style={styles.modalOverlay} 
          activeOpacity={1} 
          onPress={() => setShowTTLSelector(false)}
        >
          <BlurView intensity={20} tint="dark" style={styles.selectorModalContent}>
            <Text style={styles.selectorModalTitle}>Select Expiration</Text>
            {TTL_OPTIONS.map((option) => (
              <TouchableOpacity
                key={option.value}
                style={[
                  styles.optionButton,
                  ttl === option.value && styles.selectedOption
                ]}
                onPress={() => {
                  setTtl(option.value);
                  setShowTTLSelector(false);
                }}
              >
                <Text style={[
                  styles.optionText,
                  ttl === option.value && styles.selectedOptionText
                ]}>
                  {option.label}
                </Text>
                {ttl === option.value && <CheckCircle color="#4c669f" size={20} />}
              </TouchableOpacity>
            ))}
          </BlurView>
        </TouchableOpacity>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  headerContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingTop: 60,
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
  },
  settingsButton: {
    padding: 8,
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 12,
  },
  scrollContent: {
    padding: 20,
    paddingTop: 0,
    paddingBottom: 40,
  },
  section: {
    marginTop: 24,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#fff',
    marginLeft: 8,
  },
  card: {
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderRadius: 16,
    padding: 20,
    marginBottom: 20,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  label: {
    color: '#ccc',
    fontSize: 14,
    marginBottom: 8,
    fontWeight: '600',
  },
  input: {
    backgroundColor: 'rgba(0,0,0,0.3)',
    borderRadius: 8,
    padding: 12,
    color: '#fff',
    fontSize: 16,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  selectorButton: {
    backgroundColor: 'rgba(0,0,0,0.3)',
    borderRadius: 8,
    padding: 12,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  selectorText: {
    color: '#fff',
    fontSize: 16,
  },
  pickButton: {
    marginBottom: 10,
    borderRadius: 12,
    overflow: 'hidden',
    elevation: 5,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.3,
    shadowRadius: 4,
  },
  gradientButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 16,
  },
  buttonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    marginLeft: 10,
  },
  fileList: {
    marginBottom: 20,
  },
  fileItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 12,
    borderRadius: 12,
    marginBottom: 8,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.05)',
  },
  historyItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderRadius: 12,
    marginBottom: 8,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.05)',
    backgroundColor: 'rgba(0,0,0,0.2)',
  },
  fileInfo: {
    flex: 1,
    marginRight: 10,
  },
  fileName: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '500',
  },
  fileSize: {
    color: '#aaa',
    fontSize: 12,
    marginTop: 2,
  },
  fileActions: {
    padding: 4,
  },
  historyActions: {
    flexDirection: 'row',
  },
  actionButton: {
    padding: 8,
    marginLeft: 4,
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 8,
  },
  progressBar: {
    height: 4,
    backgroundColor: 'rgba(255,255,255,0.1)',
    borderRadius: 2,
    marginTop: 6,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#4c669f',
  },
  uploadButton: {
    flexDirection: 'row',
    backgroundColor: '#4caf50',
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    elevation: 5,
    marginTop: 10,
  },
  disabledButton: {
    opacity: 0.7,
  },
  uploadButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  modalContainer: {
    flex: 1,
    justifyContent: 'center',
    padding: 20,
  },
  modalContent: {
    backgroundColor: '#1a1a1a',
    borderRadius: 20,
    padding: 24,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.5,
    shadowRadius: 20,
    elevation: 10,
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 24,
  },
  modalTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  clearButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 16,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#ff5252',
    marginTop: 24,
  },
  clearButtonText: {
    color: '#ff5252',
    fontSize: 16,
    fontWeight: '600',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center',
    padding: 20,
  },
  selectorModalContent: {
    backgroundColor: '#1a1a1a',
    borderRadius: 16,
    padding: 20,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  selectorModalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 16,
    textAlign: 'center',
  },
  optionButton: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderRadius: 12,
    marginBottom: 8,
    backgroundColor: 'rgba(255,255,255,0.05)',
  },
  selectedOption: {
    backgroundColor: 'rgba(76, 102, 159, 0.2)',
    borderColor: '#4c669f',
    borderWidth: 1,
  },
  optionText: {
    fontSize: 16,
    color: '#ccc',
  },
  selectedOptionText: {
    color: '#fff',
    fontWeight: 'bold',
  },
});
