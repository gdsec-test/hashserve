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
//   File:          MatchHashes.java
//   Description:   Sample Java Source for MatchHashes program
//
//  History:   
//    2016.10.31   adrian chandley  
//      Created java sample. 
//
/////////////////////////////////////////////////////////////////////////////////
 
import java.io.*;
import java.util.Scanner;

public class MatchHashes {   

  public static final int  HASH_SIZE              = 144;
  public static final int  NCMEC_TARGET           = 41943;


  class _robusthash {
	String[] names;
	byte[][] hashes;
       int hashcount;
  };


  public static void main(String[] args) 
  {
    int target = NCMEC_TARGET;
    boolean best = false;
    boolean test = false;
    if (args.length < 2)
    {
      System.out.println("\nUsage:\n");
      System.out.println("    java MatchHashes <Hashfile> <HashDBfile> [distance (" + target + ")] [Table]\n");
      System.out.println("Specify Table to output matches as a table.\n");
    }
    else
    {
      if (args.length > 2)
      {
        try { target = Integer.parseInt(args[2]); }
        catch (NumberFormatException e) {};
      }
      if (args.length > 3)
      {
        char fc=args[3].charAt(0);
        if (fc=='b') best = true;
        else if (fc=='B') best = true;
        else if (fc=='t') test = true;
        else if (fc=='T') test = true;
      }
      
      _robusthash hashdata = new MatchHashes().ReadHashList(args[0]);
      System.err.println("Candidate hashes to match: " + hashdata.hashcount);
      _robusthash hashdbdata = new MatchHashes().ReadHashList(args[1]);
      System.err.println("Reference hashes in database: " + hashdbdata.hashcount);
      System.err.println("Candidate hashes with a distance less than " + target + " to a reference hash will be displayed");

      int[] matchresult;

      for (int hc = 0; hc < hashdata.hashcount; hc++)
      {
        if (!test)
        {
          if (!best) {
            matchresult = FindFirstMatch(hashdata.hashes[hc], hashdbdata, target);
          }
          else 
          {
            matchresult = FindMatch(hashdata.hashes[hc], hashdbdata);
            if (matchresult[1] > target) matchresult[0] = -1;
          }
          if (matchresult[0] > -1)
          {
            System.out.println("Match found for candidate hash " + hc + "," + hashdata.names[hc] + " to reference hash " + matchresult[0] + "," + hashdbdata.names[matchresult[0]] + " with a distance of " + matchresult[1] );
          }
        }
        else  // test
        {
          matchresult = FindMatch(hashdata.hashes[hc], hashdbdata);
          System.out.println(hashdata.names[hc]+","+matchresult[1]);
        }
        matchresult = null;
      }
    }
  }

/// <summary>
///      Reads hashes from a given file of hashes.
///      File should contains lines in the format:
///			<name>,<optional>,<..>,<optional>,<comma seperated hash value>
/// </summary>
/// <param name="path">Filename containing hashes</param>
/// <returns>_robusthash containing the hashes</returns>

  private _robusthash ReadHashList(String name)
  {
    _robusthash hashdata = new _robusthash();
    hashdata.hashcount = 0;
    int count;
    try
    {
      File f = new File(name);
      LineNumberReader  lnr = new LineNumberReader(new FileReader(f));
      lnr.skip(Long.MAX_VALUE);
      count = lnr.getLineNumber() + 1;
      lnr.close();

      hashdata.names = new String[count ];
      hashdata.hashes = new byte[count][HASH_SIZE];

      String ln = null;
      String[] parts = null;
      int val, start;
      Scanner txtfile = new Scanner(f);
      count = 0;
      while(txtfile.hasNextLine())
      {
        ln = txtfile.nextLine();
        parts = ln.split(",");
        if (parts.length >= HASH_SIZE)
        {
          hashdata.names[count] = parts[0];		// first value is a string containg the hashes name
          start = parts.length - HASH_SIZE;		// last 144 values are the hash
          for (int i = 0; i < HASH_SIZE; i++)
          {
              val = Integer.parseInt( parts[start+i] );
              hashdata.hashes[count][i] = (byte) val;
          }
          count++;
        }
      }
      txtfile.close();
      hashdata.hashcount = count;
    }
    catch (IOException e) 
    {
      System.err.println("File "+name+": "+e.getMessage());
    }
    return hashdata;
  } 

/// <summary>
///      Compares 2 hashes and returns the square of the euclidean distance.
/// </summary>
/// <param name="hash1">First hash to compare</param>
/// <param name="hash1">Second hash to compare</param>
/// <returns>Square of the euclidean distance between hashes</returns>

  private static int HashDistance(byte[] hash1, byte[] hash2)
  {
    int distance = 0;
    int diff = 0;
    for (int i = 0; i < HASH_SIZE; i++)
    {
      diff = (0xFF & hash1[i]) - (0xFF & hash2[i]);
      distance += diff * diff;
    }
    return distance;
  }

/// <summary>
///      Finds the closest match for a hash in a database of hashes.
/// </summary>
/// <param name="hash">Hash to match</param>
/// <param name="hashdb">Database (array) of hashes to be searched</param>
/// <returns>Matching DB entry number, Square of the euclidean distance between hashes</returns>

  private static int[] FindMatch(byte[] hash, _robusthash hashdb)
  {
    int[] res = new int[2];   // Note: java array values are initialised to 0
    int thisdist = 0;
    int distance = 0;
    int closest = -1;

    for (int i = 0; i < hashdb.hashcount; i++)
    {
      thisdist = HashDistance(hash, hashdb.hashes[i]);
      if ((closest == -1) || (distance > thisdist)) {
        closest = i;
        distance = thisdist;
      }
    }
    res[0] = closest;
    res[1] = distance;
    return res;
  }

/// <summary>
///   Finds the first match for a hash in a database of hashes.
/// </summary>
/// <param name="hash">Hash to match</param>
/// <param name="hashdb">Database (array) of hashes to be searched</param>
/// <param name="target">Target value below which a match will be returned</param>
/// <returns>Matching DB entry number, Square of the euclidean distance between hashes</returns>

  private static int[] FindFirstMatch(byte[] hash, _robusthash hashdb, int target)
  {
    int[] res = new int[2];   // Note: java array values are initialised to 0
    res[0] = -1;
    int distance = 0;
    int diff = 0;

    for (int i = 0; i < hashdb.hashcount; i++)
    {
      distance = 0;
      for (int j = 0; j < HASH_SIZE; j++)
      {
        diff = (0xFF & hash[j]) - (0xFF & hashdb.hashes[i][j]);
        distance += diff * diff;
        if (distance > target) break;
      }
      if (distance <= target)
      {
        res[0] = i;
        res[1] = distance ;
        break;
      }
    }
    return res;
  }

}
