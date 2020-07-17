/////////////////////////////////////////////////////////////////////////////////
// 
// IMPORTANT NOTICE
// ================
// 
//   Copyright (C) 2016, 2017, Microsoft Corporation
//   All Rights Reserved.
// 
//   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS STRICTLY UNDER THE TERMS
//   OF A SEPARATE LICENSE AGREEMENT.  NO SOURCE CODE DISTRIBUTION OR SOURCE 
//   CODE DISCLOSURE RIGHTS ARE PROVIDED WITH RESPECT TO THIS SOFTWARE OR ANY 
//   DERIVATIVE WORKS THEREOF.  USE AND NON-SOURCE CODE REDISTRIBUTION OF THE 
//   SOFTWARE IS PERMITTED ONLY UNDER THE TERMS OF SUCH LICENSE AGREEMENT.  
//   THIS SOFTWARE AND ANY WORKS/MODIFICATIONS/TRANSLATIONS DERIVED FROM THIS 
//   SOFTWARE ARE COVERED BY SUCH LICENSE AGREEMENT AND MUST ALSO CONTAIN THIS 
//   NOTICE.  THIS SOFTWARE IS CONFIDENTIAL AND SENSITIVE AND MUST BE SECURED 
//   WITH LIMITED ACCESS PURSUANT TO THE TERMS OF SUCH LICENSE AGREEMENT.
// 
//   THE SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
//   INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY 
//   AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
//   COPYRIGHT HOLDER BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, 
//   EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, 
//   PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
//   OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, 
//   WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR 
//   OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF 
//   ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//   
//   Project:       PhotoDNA Robust Hashing
//   File:          PhotoDNA.java
//   Description:   Java wrapper class for PhotoDNA libraries
//  History:   
//    2016.09.13   adrian chandley  
//      Created Java Class to enable Java apps to call PhotoDNA libraries
//
/////////////////////////////////////////////////////////////////////////////////

import java.io.File;

public class PhotoDNA {   
      private native long RobustHashInitBuffer(int initialSize);
      private native void RobustHashReleaseBuffer(long buffer);
      private native void RobustHashResetBuffer(long buffer);
      private native int ComputeRobustHash(byte[] imageData, int width,
         int height, int stride, byte[] hashValue, long buffer);
      private native int ComputeRobustHashAltColor(byte[] imageData, int width,
         int height, int stride, int color_type, byte[] hashValue,
         long buffer);
      private native int ComputeRobustHashBorder(byte[] imageData, int width,
         int height, int stride, int color_type, int[] border, byte[] hashValue,
         byte[] hashValueTrimmed, long buffer);
      private native int ComputeRobustHashAdvanced(byte[] imageData, int width,
         int height, int stride, int color_type, int sub_x, int sub_y,
         int sub_w, int sub_h, int[] border, byte[] hashValue,
         byte[] hashValueTrimmed, long buffer);
      private native int ComputeShortHash(byte[] imageData, int width,
         int height, int stride, int color_type, byte[] shortHashValue,
         long buffer);
      private native int ComputeCheckArray(byte[] imageData, int width,
         int height, int stride, int color_type, byte[] checkValue,
         long buffer);
      private native int RobustHashVersion();

// Error Codes
  public static final int  S_OK	                                           = 0x0000;
  public static final int  COMPUTE_HASH_FAILED                             = 0x0200;
  public static final int  COMPUTE_HASH_BAD_ARGUMENT                       = 0x0201;
  public static final int  COMPUTE_HASH_NO_HASH_CALCULATED                 = 0x0202;
  public static final int  COMPUTE_HASH_NO_HASH_CALCULATED_IMAGE_TOO_SMALL = 0x0213;
  public static final int  COMPUTE_HASH_NO_HASH_CALCULATED_IMAGE_IS_FLAT   = 0x0214;

  public static final int  HASH_SIZE                                      = 144;
  public static final int  SHORT_HASH_SIZE                                = 64;
  public static final int  CHECK_SIZE                                     = 36;
  public static final int  NCMEC_TARGET                                   = 41943;

// Private data
  long hashBuffer  = 0;

  private void setup(){
    String dllname = "PhotoDNA";

    String osArch = System.getProperty("os.arch");
    String osData = System.getProperty("sun.arch.data.model");

    if (osArch.equals("aarch32")) {
        dllname += "arm32";
    } else if (osArch.equals("aarch64")) {
        dllname += "arm64";
    } else if (osArch.equals("arm")) {
      if (osData != null && osData.equals("32")) {
        dllname += "arm32";
      } else {
        dllname += "arm64";
      }
    } else if (osArch.equals("amd64") || osArch.equals("x86_64")) {
      dllname += "x64";
    } else {
      dllname += "x86";
    }
    if (System.getProperty("os.name").toLowerCase().startsWith("mac os x")) {
      dllname += "-osx";
    }
    if (System.getProperty("os.name").toLowerCase().startsWith("windows")) {
      dllname += ".1.72.dll";
    } else {
      dllname += ".so.1.72";
    }
    
    // Try Class Directory
    try {
      String path = new File(System.getProperty("java.class.path")).getCanonicalPath() + "/" + dllname;
      System.load ( path ) ;
    } catch (Throwable  ex1) {
      try {       // Try Library Directory
        String path2 = "";
        String[] libPath = System.getProperty("java.library.path").split(":");
        // Loop through all paths in java.library.path to see if the lib file exists there
        for (int ii = 0; ii < libPath.length; ii++) {
          path2 = libPath[ii] + "/" + dllname;
          File tempFile = new File(path2);
          if (tempFile.exists()) {
            break;
          }
        }
        System.load ( path2 ) ;
      } catch (Throwable ex2) {
        ex2.printStackTrace();
        try {     // Try Current Directory
          String path3 = new File(".").getCanonicalPath() + "/" + dllname;
          System.load ( path3 ) ;
        } catch (Throwable ex3) {
          ex3.printStackTrace();
        }
      }
    }
  }

  PhotoDNA(){
    setup();
    hashBuffer = RobustHashInitBuffer(0);
  }
  
  PhotoDNA(int initialSize){
    setup();
    hashBuffer = RobustHashInitBuffer(initialSize);
  }
  public int ComputeHash(byte[] imageData, int width, int height, int stride,
       byte[] hashValue)
  {
    return ComputeRobustHash(imageData, width, height, stride, hashValue,
         hashBuffer);
  }
  public int ComputeHashAltColor(byte[] imageData, int width, int height,
       int stride, int color_type, byte[] hashValue)
  {
    return ComputeRobustHashAltColor(imageData, width, height, stride,
         color_type, hashValue, hashBuffer);
  }
  public int ComputeHashBorder(byte[] imageData, int width, int height,
       int stride, int color_type, int[] border, byte[] hashValue,
       byte[] hashValueTrimmed)
  {
    return ComputeRobustHashBorder(imageData, width, height, stride, color_type,
         border, hashValue, hashValueTrimmed, hashBuffer);
  }
  public int ComputeHashAdvanced(byte[] imageData, int width, int height,
       int stride, int color_type, int sub_x, int sub_y, int sub_w, int sub_h,
       int[] border, byte[] hashValue, byte[] hashValueTrimmed)
  {
    return ComputeRobustHashAdvanced(imageData, width, height, stride,
         color_type, sub_x, sub_y, sub_w, sub_h, border, hashValue,
         hashValueTrimmed, hashBuffer);
  }
  public int ComputeShort(byte[] imageData, int width, int height, int stride,
       int color_type, byte[] shortHashValue)
  {
    return ComputeShortHash(imageData, width, height, stride, color_type,
         shortHashValue, hashBuffer);
  }
  public int ComputeCheck(byte[] imageData, int width, int height, int stride,
       int color_type, byte[] checkValue)
  {
    return ComputeCheckArray(imageData, width, height, stride, color_type,
         checkValue, hashBuffer);
  }
  public int HashVersion()
  {
    return RobustHashVersion();
  }

  @Override
  protected void finalize() throws Throwable {
      try{
          RobustHashReleaseBuffer(hashBuffer);
      }catch(Throwable t){
          throw t;
      }finally{
          super.finalize();
      }
  }
}
