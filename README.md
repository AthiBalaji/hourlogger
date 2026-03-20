

A ClI tool to time your tasks, categorise them into types and write notes relating to them. 

The notes get stored in this format based on the given base_directory

```base_directory/year/month/taskname.txt```



Steps to run 
1)

```bash
hour setup
```
Enter the base directory where you want to store the notes 

2)

```bash
hour start
```
 Enter task name and type, start typing notes and end session with ':end'

```bash
:end
```
The notes will be store in the base directory in this format

base_directory/Year/Month/taskname+uniquestring


3)

```bash
hour report
```
This will give you total time logged, tasks worked on by frequency, Time spent by the type of tasks for the last 7 days. By default the timerange is last 7 days but Daterange can be given in the below format to filter dates.  

```bash
hour report YYYY-MM-DD YYYY-MM-DD
```

If only one argument for date is provided, every task up until the date is counted 
```bash
hour report YYYY-MM-DD YYYY-MM-DD
```

 Installation

Download Binary

Download the executable from the Releases section for Windows

or

Build from Source

```bash
go build -o hour.exe
```

File Structure

```
<base_path>/
  └── 2026/
        └── March/
              └── coding-20260320-153045.txt

metadata.json
```

---

Example Output for report c

```
Total Time: 5h 30m

Tasks by frequency:
- coding: 3
- debugging: 2

Time by type:
- backend: 4h
- learning: 1h 30m
```

---

